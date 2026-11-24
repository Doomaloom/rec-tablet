#!/usr/bin/env python3
"""Build the report-card curriculum catalog and its migration from the previous catalog.

Source of truth: src/workflows/swimming/curriculum/reportcards/<Level>.json, transcribed from the Lifesaving Society
report card scans (skills and tick boxes in card order, with the old subskills each box inherits marks from).
Levels without a card (Splash Fitness, Splash Private) are copied unchanged from the previous catalog.

Writes, in src/workflows/swimming/curriculum/:
  catalog.json                      the new catalog, version reportcards-<sha12 of the transcriptions>
  catalog-<previous version>.json   the previous catalog, kept so old attempts stay readable (written once)
  catalog-migration.json            how old-version attempts translate into the new version:
    skills:    new skill id -> old skill ids whose attempts count for it
    subskills: new subskill id -> {from: old subskill ids that must all be successful, fromSkillSuccess: an old fully
               successful attempt marks it successful (boxes with no old subskill)}
usage: scripts/generate-reportcard-catalog.py [--previous <catalog.json>]
"""
import hashlib, json, os, sys

ROOT = os.path.normpath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..'))
DIR = os.path.join(ROOT, 'src', 'workflows', 'swimming', 'curriculum')
CARDS = os.path.join(DIR, 'reportcards')
PREVIOUS_VERSION = 'templates-b50c7f987e101b7e'


def dump(path, data):
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write('\n')


previous_path = os.path.join(DIR, f'catalog-{PREVIOUS_VERSION}.json')
if '--previous' in sys.argv:
    previous_path = sys.argv[sys.argv.index('--previous') + 1]
if not os.path.exists(previous_path):
    # First run: the bundled catalog is still the previous version; keep a copy of it.
    current = json.load(open(os.path.join(DIR, 'catalog.json'), encoding='utf-8'))
    if current['version'] != PREVIOUS_VERSION:
        sys.exit(f'Missing {previous_path} and catalog.json is not {PREVIOUS_VERSION}')
    dump(previous_path, current)
previous = json.load(open(previous_path, encoding='utf-8'))
assert previous['version'] == PREVIOUS_VERSION, previous['version']

files = sorted(f for f in os.listdir(CARDS) if f.endswith('.json'))
raw = {f: open(os.path.join(CARDS, f), 'rb').read() for f in files}
digest = hashlib.sha256(b''.join(f.encode() + b'\0' + raw[f] for f in files)).hexdigest()
version = f'reportcards-{digest[:12]}'
cards = {json.loads(raw[f])['level']: (f, json.loads(raw[f])) for f in files}

levels, skill_map, subskill_map = [], {}, {}
for old in previous['levels']:
    if old['id'] not in cards:
        levels.append(old)
        for s in old['skills']:
            skill_map[s['id']] = [s['id']]
            for x in s['subskills']:
                subskill_map[x['id']] = {'from': [x['id']], 'fromSkillSuccess': False}
        continue
    name, card = cards.pop(old['id'])
    skills = []
    for s in card['skills']:
        skills.append({'id': s['id'], 'number': s['number'], 'name': s['name'],
                       'sourceSkillId': (s.get('fromSkills') or [s['id']])[0], 'sourceName': s['name'],
                       'sourceLines': [x['text'] for x in s['subskills']],
                       'subskills': [{'id': x['id'], 'text': x['text'], 'group': []} for x in s['subskills']]})
        skill_map[s['id']] = list(s.get('fromSkills') or [])
        for x in s['subskills']:
            subskill_map[x['id']] = {'from': list(x.get('from') or []), 'fromSkillSuccess': bool(x.get('fromSkillSuccess'))}
    levels.append({'id': old['id'], 'name': old['name'], 'source': f'reportcards/{name}',
                   'sourceSha256': hashlib.sha256(raw[name]).hexdigest(), 'skills': skills, 'kind': old['kind']})
if cards:
    sys.exit(f'Report card files for unknown levels: {sorted(cards)}')

ids = [x['id'] for l in levels for s in l['skills'] for x in s['subskills']]
assert len(ids) == len(set(ids)), 'duplicate subskill ids'
for l in levels:
    for s in l['skills']:
        assert s['subskills'] or l['kind'] != 'curriculum', s['id']
        for x in s['subskills']:
            assert x['id'].startswith(s['id'] + ':'), x['id']

catalog = {k: v for k, v in previous.items() if k != 'levels'}
catalog.update(version=version, levels=levels)
dump(os.path.join(DIR, 'catalog.json'), catalog)
dump(os.path.join(DIR, 'catalog-migration.json'), {'from': PREVIOUS_VERSION, 'to': version, 'skills': skill_map, 'subskills': subskill_map})
print(version, len(levels), 'levels', sum(len(l['skills']) for l in levels), 'skills', len(ids), 'subskills')
