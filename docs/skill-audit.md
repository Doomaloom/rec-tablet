# Attendance template skill audit

Audited 2026-10-20 against the 23 historical templates in [`COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates`](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates).

**Source rule: always use the front page for skill names and printed skill numbers, then the back page for subskills.** Back-page combined headings never replace or expand front-page skill names. Source wording, spelling, and punctuation are preserved; HTML whitespace is collapsed.

**353 front-page skills, 1,027 assigned subskills, and two explicitly unmatched back-page requirements.** SplashPrivate lists reference levels instead of its own skills. The dynamic SkillMarking sheet contains no curriculum. Both UIs use this one catalog.

## Names, numbering, and matching

- Each front marking column becomes one skill, including floats, glides, kicks, sprints, entries, surface dives, and Water Smart messages.
- Splash 7 uses 8a/8b/8c and Splash 8 uses 7a/7b/7c. Splash 9 actually prints 9 on all three stroke columns; those numbers remain 9, with separate internal IDs.
- Source group paths select the matching stroke, position, or Water Smart message. Shared distance, time, recovery, and participation requirements are retained on relevant columns. Shared back-page wording is quoted verbatim and should be read in the context of the front-page skill.
- TeenAdult3:7 remains one “Eggbeater Kick on Back or Scissor Kick on Side” skill because that is one front-page column.
- Existing internal component IDs remain stable where the skill still corresponds to one front column. Displayed numbers always come from the front page.
- Each skill stores its back-page source ID, heading, and source lines for traceability. The catalog can now be regenerated directly, without manual split overrides.

## Subskill corrections

- Restored TeenAdult2’s back-crawl flutter-kick criterion after a malformed bullet had promoted it to a heading.
- Restored TeenAdult3’s “Check the ice” criterion, which repeats its group heading.
- Separated Splash8:12a’s “Selection of appropriate procedures:” from the spoken choking question.
- Matched Splash2B’s two entry columns to the corresponding entry criteria, with the PFD-fit criterion shared. Its front-page 8c maps to the second back-page skill numbered 8.

## Source discrepancies for program review

These are in the supplied templates. The front-page names remain authoritative; the original back-page requirements remain quoted. No requirement is assigned to an unrelated front skill to hide a mismatch.

- TeenAdult2’s back-page “Within Arms’ Reach” pool-gate requirement has no corresponding front-page column.
- TeenAdult3’s back-page “Swim to Survive” message has no corresponding front-page column. The front’s 14d is “Within Arms”, which maps to the back’s “Within Arms’ Reach” group.
- Splash9’s front uses “y.” for yards; its back uses “yd”. Compact labels use yd.
- Splash8’s front medley says “(3 min)”; its back requires the medley at least three times and gives no three-minute limit. Both are preserved in their respective source roles.
- TeenAdult2’s front interval skill 16 says 15–20 second rests; its back heading says 10–15 second rests.
- Splash6’s front workout title is truncated. The complete back heading and its sets are retained as source metadata.
- TeenAdult2’s front 14a says “Front Crawl Crawl”; Splash2A’s front interval title says “utter kick”. Full names preserve these source typos; compact titles use clear stroke/kick wording.
- Other original spelling, case, and punctuation differences are retained.

## Short forms

The complete list below includes all 353 effective short forms. They identify the front-page skill; the full name and back-page checklist supply omitted assistance, depth, rests, sequence, and workout details. They are navigation labels, not standalone assessment instructions.

PFD = personal flotation device; aided = assistance/buoyant aid as specified; F/B = front/back; br. = breaststroke; EMS = emergency medical services; m = metres; yd = yards; s = seconds; min = minutes. × means repetitions. A slash between units indicates alternatives; between actions it separates actions or choices. Timed labels pair distance with its maximum time; 2:40 and 5:20 mean minutes:seconds.

## Validation

All 15 curriculum tests pass, including complete source parity, stroke/position/message matching, unmatched-source preservation, and the existing 272px compact-title measurement. An independent HTML parser also confirms that all 353 names and numbers match the front columns exactly. `go vet` passes for curriculum. The full `go test -tags ci ./cmd/... ./src/... ./next/...` run passes all packages except five pre-existing large-assessment UI tests in `next/integration`; those five failures were reproduced with the original curriculum files.

## Database behavior

Bundled version: `templates-b50c7f987e101b7e`. Existing initialization resets assessment records when the bundled version changes; roster and attendance remain. No user database was opened.

## Complete front-page skill list with back-page subskills

## Little Splash 1

Source: [LittleSplash1.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/LittleSplash1.html)

SHA-256: `39f1cdb0ff2e1f942c5e60c22fadf233133a2e2c28342ca2e8a5a1cd45e8436a`

### 1. Enter and Exit Shallow Water (Assisted)

App ID: `LittleSplash1:1` · Short form: **Assisted entry & exit**

Back-page source: `LittleSplash1:1` — Enter and Exit Shallow Water (Assisted)

1. Foot-first entry
2. Safe movement and control during entry and exit

### 2. Jump into Chest-deep water (Assisted)

App ID: `LittleSplash1:2` · Short form: **Assisted chest-deep jump**

Back-page source: `LittleSplash1:2` — Jump into Chest-deep Water (Assisted)

1. Foot-first entry
2. Balance recovered following entry

### 3. Face in Water

App ID: `LittleSplash1:3` · Short form: **Face in Water**

Back-page source: `LittleSplash1:3` — Face in water

1. Face fully submerged

### 4. Blow Bubbles in Water

App ID: `LittleSplash1:4` · Short form: **Blow Bubbles in Water**

Back-page source: `LittleSplash1:4` — Blow Bubbles in Water

1. Controlled exhalation underwater

### 5a. Float on Front Assisted (3 Seconds)

App ID: `LittleSplash1:5a` · Short form: **Aided front float 3s**

Back-page source: `LittleSplash1:5` — Float on Front and Back (3 seconds each)

1. Relaxed float on front and on back (with assistance)

### 5b. Float on Back Assisted (3 Seconds)

App ID: `LittleSplash1:5b` · Short form: **Aided back float 3s**

Back-page source: `LittleSplash1:5` — Float on Front and Back (3 seconds each)

1. Relaxed float on front and on back (with assistance)

### 6. Safe Movement in Shallow Water Wearing PFD

App ID: `LittleSplash1:6` · Short form: **PFD shallow movement**

Back-page source: `LittleSplash1:6` — Safe Movement in Shallow Water Wearing PFD

1. Moves safely through water forwards, backwards and sideways

### 7a. Glide on Front (3 m) assisted

App ID: `LittleSplash1:7a` · Short form: **Aided front glide 3m**

Back-page source: `LittleSplash1:7` — Glide on Front and Back (3 m each)

1. Streamlined Front and Back Glide
2. Distance Completed

### 7b. Glide on Back (3 m ) assisted

App ID: `LittleSplash1:7b` · Short form: **Aided back glide 3m**

Back-page source: `LittleSplash1:7` — Glide on Front and Back (3 m each)

1. Streamlined Front and Back Glide
2. Distance Completed

### 8a. Water Smart Messages - within Arms' Reach

App ID: `LittleSplash1:8a` · Short form: **Within Arms' Reach**

Back-page source: `LittleSplash1:8` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **within Arms' Reach:** Stay away from water unless you are with an adult

### 8b. Water Smart Messages - Wear a Lifejacket

App ID: `LittleSplash1:8b` · Short form: **Wear a Lifejacket**

Back-page source: `LittleSplash1:8` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat? Are they fastened properly?

## Little Splash 2

Source: [LittleSplash2.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/LittleSplash2.html)

SHA-256: `8b38b62e8761532d861f184da50c8541c7f4df9986a37d1bde779e71d95aa61f`

### 1. Enter and Exit Shallow Water Wearing PFD

App ID: `LittleSplash2:1` · Short form: **PFD shallow entry & exit**

Back-page source: `LittleSplash2:1` — Enter and Exit Shallow Water Wearing PFD

1. Appropriate PFD correctly donned and fastened on land (with assistance)
2. Foot-first entry with safe return to side
3. Safe exit

### 2. Jump into Chest-deep water

App ID: `LittleSplash2:2` · Short form: **Chest-deep jump**

Back-page source: `LittleSplash2:2` — Jump into Chest-deep Water

1. Foot-first entry
2. Controlled return to surface

### 3. Submerge

App ID: `LittleSplash2:3` · Short form: **Submerge**

Back-page source: `LittleSplash2:3` — Submerge

1. Entire body submerged

### 4. Submerge and Exhale 3 times

App ID: `LittleSplash2:4` · Short form: **Submerge & exhale ×3**

Back-page source: `LittleSplash2:4` — Submerge and Exhale 3 times

1. Entire body submerged
2. Controlled exhalation underwater 3 times

### 5a. Float on Front (3 Seconds each) Wearing PFD or other Buoyant Aid

App ID: `LittleSplash2:5a` · Short form: **PFD/aid front float 3s**

Back-page source: `LittleSplash2:5` — Float on Front and Back (3 sec. each) Wearing PFD or with other Buoyant Aid

1. Float on front and back
2. Recovery from front and back floats
3. Time requirement met

### 5b. Float on Back (3 Seconds each) Wearing PFD or other Buoyant Aid

App ID: `LittleSplash2:5b` · Short form: **PFD/aid back float 3s**

Back-page source: `LittleSplash2:5` — Float on Front and Back (3 sec. each) Wearing PFD or with other Buoyant Aid

1. Float on front and back
2. Recovery from front and back floats
3. Time requirement met

### 6. Roll Laterally Front to Back and Back to Front Wearing PFD

App ID: `LittleSplash2:6` · Short form: **PFD front/back rolls**

Back-page source: `LittleSplash2:6` — Roll Laterally Front and Back and Back to Front, Wearing PFD

1. Begin in front or back float position
2. **Controlled lateral rollover;:** Roll front to back
3. **Controlled lateral rollover;:** Roll back to front
4. Body remains horizontal

### 7a. Glide on Front (3 m each) Wearing PFD or with Other Buoyant Aid

App ID: `LittleSplash2:7a` · Short form: **PFD/aid front glide 3m**

Back-page source: `LittleSplash2:7` — Glide on Front and Back (3 m each) Wearing PFD or with other Buoyant Aid

1. Appropriate streamlined position for the type of glide
2. Minimum distance completed

### 7b. Glide on Back (3 m each) Wearing PFD or with Other Buoyant Aid

App ID: `LittleSplash2:7b` · Short form: **PFD/aid back glide 3m**

Back-page source: `LittleSplash2:7` — Glide on Front and Back (3 m each) Wearing PFD or with other Buoyant Aid

1. Appropriate streamlined position for the type of glide
2. Minimum distance completed

### 8. Flutter Kick on back 5 m with Buoyant Aid

App ID: `LittleSplash2:8` · Short form: **Aided back flutter 5m**

Back-page source: `LittleSplash2:8` — Flutter Kick on Back 5 m with Buoyant Aid

1. Alternate leg action drive
2. Rhythmic kick
3. Minimum distance completed

### 9a. Water Smart Messages - within Arms' Reach

App ID: `LittleSplash2:9a` · Short form: **Within Arms' Reach**

Back-page source: `LittleSplash2:9` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **within Arms' Reach:** Stay away from water unless you are with an adult

### 9b. Water Smart Messages - Wear a Lifejacket

App ID: `LittleSplash2:9b` · Short form: **Wear a Lifejacket**

Back-page source: `LittleSplash2:9` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat? Are they fastened properly?

## Little Splash 3

Source: [LittleSplash3.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/LittleSplash3.html)

SHA-256: `d48651d099db8d1ac5d32d73999f5190145a0cf79e4fd244b0250e49e7b4ff1b`

### 1. Jump Into Deep Water Wearing PFD, Return and Exit

App ID: `LittleSplash3:1` · Short form: **PFD deep jump & exit**

Back-page source: `LittleSplash3:1` — Jump into Deep Water Wearing PFD, Return and Exit

1. Appropriate PFD correctly donned and fastened on land
2. Foot-first entry
3. Controlled return to surface and safe return to side and exit

### 2. Sideways Entry Wearing PFD

App ID: `LittleSplash3:2` · Short form: **PFD sideways entry**

Back-page source: `LittleSplash3:2` — Sideways Entry Wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. Controlled return to surface after entry

### 3. Hold Breath Underwater 3 seconds

App ID: `LittleSplash3:3` · Short form: **Underwater breath hold 3s**

Back-page source: `LittleSplash3:3` — Hold Breath Underwater 3 Seconds

1. Entire body submerged
2. Time requirement met

### 4. Submerge and Exhale 5 Times

App ID: `LittleSplash3:4` · Short form: **Submerge & exhale ×5**

Back-page source: `LittleSplash3:4` — Submerge and Exhale 5 times

1. Entire body submerged
2. Controlled exhalation underwater 5 times

### 5. Recover Object from Bottom in Waist-Deep Water

App ID: `LittleSplash3:5` · Short form: **Waist-deep object recovery**

Back-page source: `LittleSplash3:5` — Recover Object from Bottom in Waist-Deep Water

1. Face in Water
2. Object Recovered with Hands and Returned to Surface

### 6. Back Float; Roll to Front; Swim 3 m

App ID: `LittleSplash3:6` · Short form: **Back float/roll/swim 3m**

Back-page source: `LittleSplash3:6` — Back Float; Roll to Front; Swim 3 m

1. Completion of Skills in a Continuous Sequence
2. Distance Requirement Completed

### 7a. Float on Front (5 Seconds)

App ID: `LittleSplash3:7a` · Short form: **Front float 5s**

Back-page source: `LittleSplash3:7` — Float on Front and Back (5 Second each)

1. Float on front with face in water
2. Time requirement met
3. Recovery from front and back floats

### 7b. Float on Back (5 Seconds)

App ID: `LittleSplash3:7b` · Short form: **Back float 5s**

Back-page source: `LittleSplash3:7` — Float on Front and Back (5 Second each)

1. Float on back with ears in water
2. Time requirement met
3. Recovery from front and back floats

### 8. Roll Laterally Front to Back and Back to Front

App ID: `LittleSplash3:8` · Short form: **Front/back rolls**

Back-page source: `LittleSplash3:8` — Roll Laterally Front to Back and Back to Front

1. Begin in front or back float position
2. **Controlled lateral rollover;:** Roll front to back
3. **Controlled lateral rollover;:** Roll back to front

### 9a. Glide on Front (3 m)

App ID: `LittleSplash3:9a` · Short form: **Front glide 3m**

Back-page source: `LittleSplash3:9` — Glide on Front and Back (3 m each)

1. **Front glide:** face in water
2. **Front glide:** arms extended beyond head
3. Minimum distance completed

### 9b. Glide on Back (3 m)

App ID: `LittleSplash3:9b` · Short form: **Back glide 3m**

Back-page source: `LittleSplash3:9` — Glide on Front and Back (3 m each)

1. **Back Glide:** ears in water
2. **Back Glide:** arms by sides
3. Minimum distance completed

### 10. Flutter Kick on Back 5 m

App ID: `LittleSplash3:10` · Short form: **Flutter Kick on Back 5 m**

Back-page source: `LittleSplash3:10` — Flutter Kick on Back 5 m

1. Body on back
2. Alternate leg action drive
3. Rhythmic kick
4. Minimum distance completed

### 11. Flutter Kick on Front 5 m

App ID: `LittleSplash3:11` · Short form: **Flutter Kick on Front 5 m**

Back-page source: `LittleSplash3:11` — Flutter Kick on Front 5 m

1. Body on front
2. Alternate leg action drive
3. Rhythmic kick
4. Minimum distance completed

### 12a. Water Smart Messages - Within Arms' Reach

App ID: `LittleSplash3:12a` · Short form: **Within Arms' Reach**

Back-page source: `LittleSplash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Stay away from water unless you are with an adult

### 12b. Water Smart Messages - Wear a Lifejacket

App ID: `LittleSplash3:12b` · Short form: **Wear a Lifejacket**

Back-page source: `LittleSplash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat?
3. **Wear a Lifejacket:** Are they fastened properly?

## Little Splash 4

Source: [LittleSplash4.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/LittleSplash4.html)

SHA-256: `093554d92fbca81692d5b5c92e89099b91776674e8f85141a12f4518695b54d6`

### 1. Jump Into Deep Water, Return and Exit

App ID: `LittleSplash4:1` · Short form: **Deep jump, return & exit**

Back-page source: `LittleSplash4:1` — Jump into Deep Water, Return and Exit

1. Foot-first entry
2. Controlled return to surface and safe return to side and exit

### 2. Sideways Entry

App ID: `LittleSplash4:2` · Short form: **Sideways Entry**

Back-page source: `LittleSplash4:2` — Sideways Entry

1. Controlled return to surface after entry

### 3. Tread Water 10 Seconds Wearing PFD

App ID: `LittleSplash4:3` · Short form: **PFD tread 10s**

Back-page source: `LittleSplash4:3` — Tread Water 10 Seconds Wearing PFD

1. Mouth and nose above surface
2. Sculling action of hand generates support
3. Time requirement met

### 4. Open Eyes Underwater

App ID: `LittleSplash4:4` · Short form: **Open Eyes Underwater**

Back-page source: `LittleSplash4:4` — Open Eyes Underwater

1. Full face submerged with eyes open

### 5. Recover Object from Bottom in Chest-Deep Water

App ID: `LittleSplash4:5` · Short form: **Chest-deep object recovery**

Back-page source: `LittleSplash4:5` — Recover Object from Bottom in Chest-Deep Water

1. Face in Water
2. Object Recovered with Hands and Returned to Surface

### 6. Wearing a PFD, Sideways Entry Into Deep Water; Tread 15 seconds; Swim/Kick 5 m

App ID: `LittleSplash4:6` · Short form: **PFD entry/tread 15s/5m**

Back-page source: `LittleSplash4:6` — Wearing a PFD, Sideways Entry into Deep Water; Tread 15 Seconds; Swim/Kick 5 m

1. Completion of Skills in a Continuous Sequence
2. Distance Requirement Completed

### 7. Front Float; Roll to Back; Swim 5 m

App ID: `LittleSplash4:7` · Short form: **Front float/roll/swim 5m**

Back-page source: `LittleSplash4:7` — Front Float; Roll to Back; Swim 5 m

1. Completion of Skills in a continuous sequence
2. Distance requirement completed

### 8. Glide on side 3 m

App ID: `LittleSplash4:8` · Short form: **Glide on side 3 m**

Back-page source: `LittleSplash4:8` — Glide on side 3 m

1. **Streamlined side glide;:** body on side;
2. **Streamlined side glide;:** bottom arm extended beyond head
3. **Streamlined side glide;:** top arm by side
4. **Streamlined side glide;:** head resting on bottom arm
5. Minimum distance completed

### 9a. Flutter Kick; On Front 7 m

App ID: `LittleSplash4:9a` · Short form: **Front flutter 7m**

Back-page source: `LittleSplash4:9` — Flutter Kick: on front 7m; on back 7m; on side 5 m

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive
3. Minimum distance completed in each position

### 9b. Flutter Kick; On Back 7 m

App ID: `LittleSplash4:9b` · Short form: **Back flutter 7m**

Back-page source: `LittleSplash4:9` — Flutter Kick: on front 7m; on back 7m; on side 5 m

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive
3. Minimum distance completed in each position

### 9c. Flutter Kick; On Side 5 m

App ID: `LittleSplash4:9c` · Short form: **Side flutter 5m**

Back-page source: `LittleSplash4:9` — Flutter Kick: on front 7m; on back 7m; on side 5 m

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive
3. Minimum distance completed in each position

### 10. Front Crawl 5 m Wearing PFD

App ID: `LittleSplash4:10` · Short form: **PFD front crawl 5m**

Back-page source: `LittleSplash4:10` — Front Crawl 5 m Wearing PFD

1. Body on front
2. Alternate arm action
3. Propulsive, rhythmic flutter kick with alternate leg drive
4. Minimum distance completed

### 11a. Water Smart Messages - Within Arms' Reach

App ID: `LittleSplash4:11a` · Short form: **Within Arms' Reach**

Back-page source: `LittleSplash4:11` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Stay away from water unless you are with an adult

### 11b. Water Smart Messages - Wear a Lifejacket

App ID: `LittleSplash4:11b` · Short form: **Wear a Lifejacket**

Back-page source: `LittleSplash4:11` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Is everone in my family wearing lifejackets in the boat?
3. **Wear a Lifejacket:** Are they fastened properly?

## Little Splash 5

Source: [LittleSplash5.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/LittleSplash5.html)

SHA-256: `8642b1c54fb1aad0cb188cf29b0458152a1e19ed3d74d94dcd0df0fa3cac27a0`

### 1. Forward Roll Entry Wearing PFD

App ID: `LittleSplash5:1` · Short form: **PFD forward roll entry**

Back-page source: `LittleSplash5:1` — Forward Roll Entry Wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. Controlled entry and return to surface

### 2. Tread Water 10 Seconds

App ID: `LittleSplash5:2` · Short form: **Tread Water 10 Seconds**

Back-page source: `LittleSplash5:2` — Tread Water 10 Seconds

1. Vertical body position
2. Mouth and nose above surface
3. Sculling action of hand generates support
4. Supportive kick
5. Minimum time met

### 3. Submerge and Hold Breath 5 Seconds

App ID: `LittleSplash5:3` · Short form: **Underwater breath hold 5s**

Back-page source: `LittleSplash5:3` — Submerge and Hold Breath 5 Seconds

1. Entire body submerged
2. Time requirement met

### 4. Recover Object from Bottom in Chest-deep Water

App ID: `LittleSplash5:4` · Short form: **Chest-deep object recovery**

Back-page source: `LittleSplash5:4` — Recover Object from Bottom in Chest-deep Water

1. Face in water and feet off the bottom
2. Object recovered with hands and returned to surface

### 5. Wearing a PFD, sideways entry into deep Water; Tread 20 Seconds; Swim/Kick 10m

App ID: `LittleSplash5:5` · Short form: **PFD entry/tread 20s/10m**

Back-page source: `LittleSplash5:5` — Wearing a PFD, Sideways Entry into Deep Water; Tread 20 Seconds; Swim/Kick 10 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 6. Whip Kick in vertical position (20 Seconds) with a PFD or Buoyant aid

App ID: `LittleSplash5:6` · Short form: **Aided vertical whip 20s**

Back-page source: `LittleSplash5:6` — Whip kick in vertical position (20 sec) with a PFD or buoyant aid

1. Kick in vertical position
2. Kick is simultaneous and symmetrical;heels recover towards buttocks
3. Legs drive with knees apart; Feet wider than knees
4. minimum time met

### 7. Front Crawl 5 m

App ID: `LittleSplash5:7` · Short form: **Front Crawl 5 m**

Back-page source: `LittleSplash5:7` — Front crawl 5 m

1. Body on front
2. Alternate arm action
3. Propulsive, rhythmic flutter kick with alternate leg drive; Slight knee bend
4. Breathing with underwater exhalation
5. Minimum distance completed

### 8. Back Crawl 5 m

App ID: `LittleSplash5:8` · Short form: **Back Crawl 5 m**

Back-page source: `LittleSplash5:8` — Back Crawl 5 m

1. Body on back
2. Alternate arm action
3. Propulsive, rhythmic flutter kick with alternate leg drive near the surface
4. Relaxed breathing
5. Minimum distance completed

### 9. Interval training: 4x5 m Flutter Kick on Back with 30 Second rests

App ID: `LittleSplash5:9` · Short form: **Back flutter 4×5m**

Back-page source: `LittleSplash5:9` — Interval Training: 4x5 m Flutter Kick on Back with 30 Seconds Rests

1. Repetitions completed
2. Propulsive, rhythmic flutter kick with alternate leg drive near the surface

### 10a. Water Smart Messages - Within Arms' Reach

App ID: `LittleSplash5:10a` · Short form: **Within Arms' Reach**

Back-page source: `LittleSplash5:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Stay away from water unless you are with an adult

### 10b. Water Smart Messages - Wear a Lifejacket

App ID: `LittleSplash5:10b` · Short form: **Wear a Lifejacket**

Back-page source: `LittleSplash5:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat?
3. **Wear a Lifejacket:** Are they fastened properly?

## Parentand Tot 1

Source: [ParentandTot1.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/ParentandTot1.html)

SHA-256: `f60bf7207dfb6a4fa6617b0902765940b62d5b5f00d80141e3352c7be0767494`

### 1. Enter and Exit the Water Safely with Tot

App ID: `ParentandTot1:1` · Short form: **Safe entry & exit with tot**

Back-page source: `ParentandTot1:1` — Enter and Exit the Water Safely with Tot

1. Foot first entry (with assistance)

### 2. Readiness for Submersion

App ID: `ParentandTot1:2` · Short form: **Readiness for Submersion**

Back-page source: `ParentandTot1:2` — Readiness for Submersion

1. Parent encouraging child to wet face
2. Child comfortable with water on face

### 3. Hold Tot on Front, Eye Contact

App ID: `ParentandTot1:3` · Short form: **Front hold & eye contact**

Back-page source: `ParentandTot1:3` — Hold Tot on Front, Eye Contact

1. Relaxed Front Position for a few seconds (with assistance)
2. Parent-child eye contact

### 4. Hold Tot on Back, Head and Back Support

App ID: `ParentandTot1:4` · Short form: **Supported back hold**

Back-page source: `ParentandTot1:4` — Hold Tot on Back, Head and Back Support

1. Relaxed Back Position (with assistance)

### 5. Front Float (Face Out) - Assisted

App ID: `ParentandTot1:5` · Short form: **Assisted front float**

Back-page source: `ParentandTot1:5` — Relaxed Front Float (with assistance)

1. Relaxed front position for a few seconds (with assistance)
2. Parent supporting child face-to-face with proper hold

### 6. Back Float (Assisted)

App ID: `ParentandTot1:6` · Short form: **Back Float (Assisted)**

Back-page source: `ParentandTot1:6` — Back Float (Assisted)

1. Relaxed back float (with assistance)
2. Parent using proper support hold

### 7. Float Wearing PFD (Assisted)

App ID: `ParentandTot1:7` · Short form: **Assisted PFD float**

Back-page source: `ParentandTot1:7` — Float Wearing PFD (Assisted)

1. Parent selecting and putting appropriate PFD on Child
2. Relaxed float in any position (with assistance)

### 8. Arms: Splashing, Reaching, Paddling, On Front and Back

App ID: `ParentandTot1:8` · Short form: **Front/back arm movements**

Back-page source: `ParentandTot1:8` — Arms: Splashing, Reaching, Paddling, On Front and Back

1. Child relaxed and comfortable with splashing
2. Child moves arms (assisted by parent)

### 9. Legs: Tickling, Splashing, Kicking, on Front and Back

App ID: `ParentandTot1:9` · Short form: **Front/back leg movements**

Back-page source: `ParentandTot1:9` — Legs: Tickling, Splashing, Kicking, On Front and Back

1. Relaxed, alternate leg action by child (assisted by parent)
2. On front - chin in water; on back - ears in water
3. Child comfortable with splashing

### 10a. Water Smart Messages - Within Arms' Reach

App ID: `ParentandTot1:10a` · Short form: **Within Arms' Reach**

Back-page source: `ParentandTot1:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Designate a backyard pool lifeguard

### 10b. Water Smart Messages - Wear a Lifejacket

App ID: `ParentandTot1:10b` · Short form: **Wear a Lifejacket**

Back-page source: `ParentandTot1:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Most parents make children wear lifejackets but don't wear one themselves

### 10c. Water Smart Messages - Swim to Survive

App ID: `ParentandTot1:10c` · Short form: **Swim to Survive**

Back-page source: `ParentandTot1:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Swim to Survive:** Swim lessons are the first step to being safe around water. Take lessons until the end of the curriculum.

## Parentand Tot 2

Source: [ParentandTot2.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/ParentandTot2.html)

SHA-256: `5617a535a8e8c4864c025932a974434903085613b2c9c3652b4af5b777929965`

### 1. Entry from Sitting Position (Assisted)

App ID: `ParentandTot2:1` · Short form: **Assisted seated entry**

Back-page source: `ParentandTot2:1` — Entry from Sitting Position (Assisted)

1. Foot-first entry
2. Safe movement and control during entry

### 2. Exit the Water (Assisted)

App ID: `ParentandTot2:2` · Short form: **Exit the Water (Assisted)**

Back-page source: `ParentandTot2:2` — Exit the Water (Assisted)

1. Safe movement and control during exit

### 3. Blow Bubbles On and In Water

App ID: `ParentandTot2:3` · Short form: **Bubbles on/in water**

Back-page source: `ParentandTot2:3` — Blow Bubbles On and In Water

1. Controlled exhalation on surface and underwater

### 4. Face Wet and in Water

App ID: `ParentandTot2:4` · Short form: **Face Wet and in Water**

Back-page source: `ParentandTot2:4` — Face Wet and In Water

1. Child comfortable with water on face
2. Face fully submerged

### 5. Attempt to Recover Object Below Surface

App ID: `ParentandTot2:5` · Short form: **Attempt subsurface pickup**

Back-page source: `ParentandTot2:5` — Attempt to Recover Object Below Surface

1. Attempt to recover object with hands (with Assistance)

### 6. Entry from Sitting Position Wearing PFD and Return (Assisted)

App ID: `ParentandTot2:6` · Short form: **Assisted PFD entry/return**

Back-page source: `ParentandTot2:6` — Entry From Sitting Position Wearing PFD and Return (Assisted)

1. Parent selecting and putting appropriate PFD on child
2. **Child waiting for adult to enter water first;:** Adult prepared and cueing
3. Safe entry
4. Parent initiates return to point of entry or safety and secures child

### 7. Front Float (Face In) - Assisted

App ID: `ParentandTot2:7` · Short form: **Assisted face-in front float**

Back-page source: `ParentandTot2:7` — Front Float - Face In (Assisted)

1. Relaxed front float with face in water (with Assistance)
2. Parent using proper support hold

### 8. Back Float (Assisted)

App ID: `ParentandTot2:8` · Short form: **Back Float (Assisted)**

Back-page source: `ParentandTot2:8` — Back Float (Assisted)

1. Relaxed back float (with assistance)
2. Parent using proper support hold

### 9a. Kicking on Front (Assisted)

App ID: `ParentandTot2:9a` · Short form: **Assisted front kick**

Back-page source: `ParentandTot2:9` — Kicking on Front and Back (Assisted)

1. **Streamlined body position::** On front, face in, arms extended beyond head
2. Alternate leg action (flutter kick)

### 9b. Kicking on Back (Assisted)

App ID: `ParentandTot2:9b` · Short form: **Assisted back kick**

Back-page source: `ParentandTot2:9` — Kicking on Front and Back (Assisted)

1. **Streamlined body position::** On back, ears in the water, arms by sides
2. Alternate leg action (flutter kick)

### 10. Surface Passes with Continuous Contact

App ID: `ParentandTot2:10` · Short form: **Contact surface passes**

Back-page source: `ParentandTot2:10` — Surface Passes with Continuous Contact

1. Child in continuous contact with either parent or instructor
2. Child relaxed with chin at surface or in water

### 11a. Water Smart Messages - Within Arms' Reach

App ID: `ParentandTot2:11a` · Short form: **Within Arms' Reach**

Back-page source: `ParentandTot2:11` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Stay tub-side

### 11b. Water Smart Messages - Wear a Lifejacket

App ID: `ParentandTot2:11b` · Short form: **Wear a Lifejacket**

Back-page source: `ParentandTot2:11` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Most drowning victims knew how to swim...wear your lifejacket

### 11c. Water Smart Messages - Swim to Survive

App ID: `ParentandTot2:11c` · Short form: **Swim to Survive**

Back-page source: `ParentandTot2:11` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Swim to Survive:** Swimming skills need to be taught

## Parentand Tot 3

Source: [ParentandTot3.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/ParentandTot3.html)

SHA-256: `c5c5200d4266167b4d28b7acb084e29e8f339681df55efe1150baac9d5866d0a`

### 1. Jump Entry (Assisted)

App ID: `ParentandTot3:1` · Short form: **Jump Entry (Assisted)**

Back-page source: `ParentandTot3:1` — Jump Entry (Assisted)

1. Foot-first entry
2. Safe movement and control during entry

### 2. Entry and Submerge from Sitting Position (Assisted)

App ID: `ParentandTot3:2` · Short form: **Assisted seated submersion**

Back-page source: `ParentandTot3:2` — Entry and Submerge from Sitting Position (Assisted)

1. Head or face submerged upon entry
2. Safe movement and control during entry

### 3. Exit the Water (Unassisted)

App ID: `ParentandTot3:3` · Short form: **Unassisted exit**

Back-page source: `ParentandTot3:3` — Exit the Water (Unassisted)

1. Safe movement and control during exit

### 4. Hold Breath Underwater (Assisted)

App ID: `ParentandTot3:4` · Short form: **Assisted breath hold**

Back-page source: `ParentandTot3:4` — Hold Breath Underwater (Assisted)

1. Child holding breath when face underwater

### 5. Attempt to Open Eyes Underwater

App ID: `ParentandTot3:5` · Short form: **Try eyes open underwater**

Back-page source: `ParentandTot3:5` — Attempt to Open Eyes Underwater

1. Face fully submerged

### 6. Attempt to Recover Object From Bottom

App ID: `ParentandTot3:6` · Short form: **Attempt bottom pickup**

Back-page source: `ParentandTot3:6` — Attempt to Recover Object from Bottom

1. Attempt to recover object with hands (with assistance)
2. Face in water

### 7. Standing Jump Entry, return to Edge (Assisted)

App ID: `ParentandTot3:7` · Short form: **Assisted jump & return**

Back-page source: `ParentandTot3:7` — Standing Jump Entry; Return to Edge (Assisted)

1. Child waiting for adult to enter water first; Adult prepared and cueing
2. Safe entry
3. Parent initiates return to point of entry or safety and secures child

### 8. Jump Entry and Float Wearing PFD (Assisted)

App ID: `ParentandTot3:8` · Short form: **Assisted PFD jump/float**

Back-page source: `ParentandTot3:8` — Jump Entry and Float Wearing PFD (Assisted)

1. Child donning appropriate PFD on land (with assistance)
2. Child waits for adult to enter water first; Adult prepared and cueing
3. Safe, foot-first entry; Child orients self then begins float
4. Relaxed front or back float

### 9a. Front "Starfish" Floats (Assisted)

App ID: `ParentandTot3:9a` · Short form: **Aided front starfish float**

Back-page source: `ParentandTot3:9` — Front and Back "Starfish" Floats (Assisted)

1. Relaxed front and back starfish float (with assistance)

### 9b. Back "Starfish" Floats (Assisted)

App ID: `ParentandTot3:9b` · Short form: **Aided back starfish float**

Back-page source: `ParentandTot3:9` — Front and Back "Starfish" Floats (Assisted)

1. Relaxed front and back starfish float (with assistance)

### 10a. Front "Pencil" Floats (Assisted)

App ID: `ParentandTot3:10a` · Short form: **Aided front pencil float**

Back-page source: `ParentandTot3:10` — Front and Back "Pencil" Floats (Assisted)

1. Relaxed front and back pencil floats (with assistance)

### 10b. Back "Pencil" Floats (Assisted)

App ID: `ParentandTot3:10b` · Short form: **Aided back pencil float**

Back-page source: `ParentandTot3:10` — Front and Back "Pencil" Floats (Assisted)

1. Relaxed front and back pencil floats (with assistance)

### 11a. Kicking on Front (Assisted)

App ID: `ParentandTot3:11a` · Short form: **Assisted front kick**

Back-page source: `ParentandTot3:11` — Kicking on Front and Back (Assisted)

1. **Streamlined body position;:** on Front, Face in, arms extended beyond head
2. Alternate leg action (flutter kick)

### 11b. Kicking on Back (Assisted)

App ID: `ParentandTot3:11b` · Short form: **Assisted back kick**

Back-page source: `ParentandTot3:11` — Kicking on Front and Back (Assisted)

1. **Streamlined body position;:** on Back, Ears in water, arms by sides
2. Alternate leg action (flutter kick)

### 12. Underwater Passes

App ID: `ParentandTot3:12` · Short form: **Underwater Passes**

Back-page source: `ParentandTot3:12` — Underwater Passes

1. Child in continuous contact with either parent or instructor
2. Child relaxed
3. Head and body submerged

### 13a. Water Smart Messages - Within Arms' Reach

App ID: `ParentandTot3:13a` · Short form: **Within Arms' Reach**

Back-page source: `ParentandTot3:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Within Arms' Reach:** Drain bathtubs and standing water

### 13b. Water Smart Messages - Wear a Lifejacket

App ID: `ParentandTot3:13b` · Short form: **Wear a Lifejacket**

Back-page source: `ParentandTot3:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Wear a Lifejacket:** Lifejackets do not replace supervision by an adult

### 13c. Water Smart Messages- Swim to Survive

App ID: `ParentandTot3:13c` · Short form: **Swim to Survive**

Back-page source: `ParentandTot3:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart Messages
2. **Swim to Survive:** Swim to survive teaches the minimum basic skills required to survive an unexpected fall into deep water

## Splash 1

Source: [Splash1.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash1.html)

SHA-256: `aa217392103b8f9640f81f1ab5d1cf6fc4729f46354f4bcb7edc5aa7ad0c2827`

### 1. Enter and Exit Shallow Water

App ID: `Splash1:1` · Short form: **Shallow entry & exit**

Back-page source: `Splash1:1` — Enter and Exit Shallow Water

1. Foot-first entry
2. Safe movement and control during entry and exit

### 2. Jump into Chest-deep water

App ID: `Splash1:2` · Short form: **Chest-deep jump**

Back-page source: `Splash1:2` — Jump into Chest-deep Water

1. Foot-first entry and controlled return to surface

### 3. Tread water (30 sec) wearing PFD

App ID: `Splash1:3` · Short form: **PFD tread 30s**

Back-page source: `Splash1:3` — Tread water (30 sec.) wearing PFD

1. Vertical body position
2. Mouth and nose above surface
3. Sculling action of hand generates support
4. Continuous kick
5. Minimum time met

### 4. Hold breath underwater 5 sec.

App ID: `Splash1:4` · Short form: **Underwater breath hold 5s**

Back-page source: `Splash1:4` — Hold breath underwater 5 sec.

1. Entire body submerged
2. Minimum time met

### 5. Submerge and exhale 5 times

App ID: `Splash1:5` · Short form: **Submerge & exhale ×5**

Back-page source: `Splash1:5` — Submerge and exhale 5 times

1. Entire body submerged
2. Controled exhalation underwater 5 times

### 6a. Float on front 5 sec.

App ID: `Splash1:6a` · Short form: **Front float 5s**

Back-page source: `Splash1:6` — Float on front and back (5 sec. each)

1. Float on front with face in water
2. Recovery from front and back float
3. Minimum time met

### 6b. Float on back 5 sec.

App ID: `Splash1:6b` · Short form: **Back float 5s**

Back-page source: `Splash1:6` — Float on front and back (5 sec. each)

1. Float on back with ears in water
2. Recovery from front and back float
3. Minimum time met

### 7. Roll laterally front to back and back to front

App ID: `Splash1:7` · Short form: **Front/back rolls**

Back-page source: `Splash1:7` — Roll laterally front to back and back to front

1. Begin in front or back float position
2. Controlled lateral rollover; Roll front to back; Roll back to front
3. Body remains horizontal

### 8a. Glide on front 3 m

App ID: `Splash1:8a` · Short form: **Front glide 3m**

Back-page source: `Splash1:8` — Glide on front, back and side (3 m each)

1. Front glide – face in water; underwater exhalation; arms extended beyond head
2. Minimum distance completed

### 8b. Glide on back 3 m

App ID: `Splash1:8b` · Short form: **Back glide 3m**

Back-page source: `Splash1:8` — Glide on front, back and side (3 m each)

1. Back glide – ears in waters; arms by sides
2. Minimum distance completed

### 8c. Glide on side 3 m

App ID: `Splash1:8c` · Short form: **Side glide 3m**

Back-page source: `Splash1:8` — Glide on front, back and side (3 m each)

1. Side glide – body on side; bottom arm extended beyond head; top arm by side; head resting on bottom arm
2. Minimum distance completed

### 9a. Flutter kick on front 5 m

App ID: `Splash1:9a` · Short form: **Front flutter 5m**

Back-page source: `Splash1:9` — Flutter kick on front and back (5 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive
3. Minimum distance completed

### 9b. Flutter kick on back 5 m

App ID: `Splash1:9b` · Short form: **Back flutter 5m**

Back-page source: `Splash1:9` — Flutter kick on front and back (5 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive
3. Minimum distance completed

### 10a. Front crawl 5 m wearing PFD

App ID: `Splash1:10-front-crawl-wearing-pfd-5-m` · Short form: **PFD front crawl 5m**

Back-page source: `Splash1:10` — Front Crawl and Back Crawl wearing PFD (5 m)

1. **Front Crawl wearing PFD (5 m):** Body on front
2. **Front Crawl wearing PFD (5 m):** Alternate arm action
3. **Front Crawl wearing PFD (5 m):** Propulsive, rhythmic flutter kick with alternate leg drive
4. **Front Crawl wearing PFD (5 m):** Minimum distance completed

### 10b. Back crawl 5 m wearing PFD

App ID: `Splash1:10-back-crawl-wearing-pfd-5-m` · Short form: **PFD back crawl 5m**

Back-page source: `Splash1:10` — Front Crawl and Back Crawl wearing PFD (5 m)

1. **Back Crawl wearing PFD (5 m):** Body on back
2. **Back Crawl wearing PFD (5 m):** Alternate arm action
3. **Back Crawl wearing PFD (5 m):** Propulsive, rhythmic flutter kcik with alternate leg drive

### 11a. Water Smart Messages: Swim with a Buddy

App ID: `Splash1:11a` · Short form: **Swim with a Buddy**

Back-page source: `Splash1:11` — Water Smart Messages:

1. Participation in a water activity reinforcing the Water Smart messages
2. **Swim with a Buddy:** Never swim alone, always with a buddy

### 11b. Water Smart Messages: Wear a Lifejacket

App ID: `Splash1:11b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash1:11` — Water Smart Messages:

1. Participation in a water activity reinforcing the Water Smart messages
2. **Wear a Lifejacket:** is everyone in my family wearing lifejackets in the boat Including me? Mom?Dad?Others?

### 11c. Water Smart Messages: Check the Ice

App ID: `Splash1:11c` · Short form: **Check the Ice**

Back-page source: `Splash1:11` — Water Smart Messages:

1. Participation in a water activity reinforcing the Water Smart messages
2. **Check the Ice:** Always have an adult check the ice thickness before going on it

### 11d. Water Smart Messages: Swim to Survive

App ID: `Splash1:11d` · Short form: **Swim to Survive**

Back-page source: `Splash1:11` — Water Smart Messages:

1. Participation in a water activity reinforcing the Water Smart messages
2. **Swim to Survive:** Follow the bubbles to the surface

## Splash 2A

Source: [Splash2A.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash2A.html)

SHA-256: `d354b0e45d44e0811d7896df0b44357857cdd92838ccaf1dc4ff60fbd121893b`

### 1. Jump into deep water wearing PFD

App ID: `Splash2A:1` · Short form: **PFD deep-water jump**

Back-page source: `Splash2A:1` — Jump into deep water wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. Foot-first entry and controlled return to surface

### 2. Tread water 10 sec

App ID: `Splash2A:2` · Short form: **Tread water 10 sec**

Back-page source: `Splash2A:2` — Tread water 10 sec

1. Vertical body position
2. Mouth and nose above surface
3. Sculling action of hand generates support
4. Supportive kick
5. Minimum time met

### 3. Open eyes underwater

App ID: `Splash2A:3` · Short form: **Open eyes underwater**

Back-page source: `Splash2A:3` — Open eyes underwater

1. Full face submerged with eyes open

### 4. Recover object from bottom in chest-deep water

App ID: `Splash2A:4` · Short form: **Chest-deep object recovery**

Back-page source: `Splash2A:4` — Recover object from bottom in chest-deep water

1. Face in water
2. Object recovered with hands and returned to surface

### 5. Wearing PFD, jump into deep water, tread 20 sec. and swim/ kick 10 m

App ID: `Splash2A:5` · Short form: **PFD jump/tread 20s/10m**

Back-page source: `Splash2A:5` — Wearing PFD, jump into deep water, tread 20 sec, and swim/kick 10 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 6a. Flutter kick on front 7 m

App ID: `Splash2A:6a` · Short form: **Front flutter 7m**

Back-page source: `Splash2A:6` — Flutter kick on Front, Back and Side (7 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 6b. Flutter kick on back 7 m

App ID: `Splash2A:6b` · Short form: **Back flutter 7m**

Back-page source: `Splash2A:6` — Flutter kick on Front, Back and Side (7 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 6c. Flutter kick on side 7 m

App ID: `Splash2A:6c` · Short form: **Side flutter 7m**

Back-page source: `Splash2A:6` — Flutter kick on Front, Back and Side (7 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 7. Front kick 3 sec. to side kick 3 sec. to front kick 3 sec. – assisted

App ID: `Splash2A:7` · Short form: **Aided front/side kick 3s**

Back-page source: `Splash2A:7` — Front kick 3 sec. to side kick 3 sec. to front kick 3 sec. – assisted

1. *Note: Can be instructor-assisted or while holding buoyant aid
2. Completion of skills in a continuous sequence
3. Appropriate streamlined body position
4. Propulsive, rhythmic flutter kick with alternate leg-drive; slight knee bend
5. Time met in each body position

### 8a. Front crawl 5 m

App ID: `Splash2A:8-front-crawl` · Short form: **Front crawl 5m**

Back-page source: `Splash2A:8` — Front Crawl and Back Crawl (5 m)

1. Minimum distance completed (for both strokes)
2. **Front Crawl:** Body on front
3. **Front Crawl:** Alternate arm action
4. **Front Crawl:** Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
5. **Front Crawl:** Breathing pattern with underwater exhalation

### 8b. Back crawl 5 m

App ID: `Splash2A:8-back-crawl` · Short form: **Back crawl 5m**

Back-page source: `Splash2A:8` — Front Crawl and Back Crawl (5 m)

1. Minimum distance completed (for both strokes)
2. **Back crawl:** Body on back
3. **Back crawl:** Alternate arm action
4. **Back crawl:** Propulsive, rhythmic flutter kick with alternate leg drive near the surface
5. **Back crawl:** Relaxed breathing
6. **Back crawl:** Minimum distance completed

### 9. Interval training: 4 x 5 m utter kick with 20 sec. rests

App ID: `Splash2A:9` · Short form: **Flutter kick 4×5m**

Back-page source: `Splash2A:9` — Interval training: 4x5 m flutter kick with 20 sec rests

1. Repetitions completed
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend

### 10a. Water Smart Messages: Swim with a Buddy

App ID: `Splash2A:10a` · Short form: **Swim with a Buddy**

Back-page source: `Splash2A:10` — Water Smart Messages :

1. Participation in a water activity reinforcing the water smart message
2. **Swim with a buddy:** Never swim alone, always with a buddy

### 10b. Water Smart Messages: Wear a Lifejacket

App ID: `Splash2A:10b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash2A:10` — Water Smart Messages :

1. Participation in a water activity reinforcing the water smart message
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 10c. Water Smart Messages: Check the Ice

App ID: `Splash2A:10c` · Short form: **Check the Ice**

Back-page source: `Splash2A:10` — Water Smart Messages :

1. Participation in a water activity reinforcing the water smart message
2. **Check the Ice:** Always have an adult check ice thickness before going on it

### 10d. Water Smart Messages: Swim to Survive

App ID: `Splash2A:10d` · Short form: **Swim to Survive**

Back-page source: `Splash2A:10` — Water Smart Messages :

1. Participation in a water activity reinforcing the water smart message
2. **Swim to Survive:** Follow the bubbles to the surface

## Splash 2B

Source: [Splash2B.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash2B.html)

SHA-256: `2132d518a552db46a13cba8f86b3f0573d657a09569dc8f76d625831800f181c`

### 1. Jump into deep water, return and exit

App ID: `Splash2B:1` · Short form: **Deep jump, return & exit**

Back-page source: `Splash2B:1` — Jump into deep water, return and exit

1. Foot-first entry
2. Controlled return to surface and safe return to side

### 2a. Sideways entry wearing PFD

App ID: `Splash2B:2a` · Short form: **PFD sideways entry**

Back-page source: `Splash2B:2` — Sideways entry wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. controlled return to surface after entry

### 2b. Forward roll entry wearing PFD

App ID: `Splash2B:2b` · Short form: **PFD forward roll entry**

Back-page source: `Splash2B:2` — Sideways entry wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. **Forward roll entry wearing PFD:** Appropriate Controlled entry and return to surface

### 3. Tread water 30 sec

App ID: `Splash2B:3` · Short form: **Tread water 30 sec**

Back-page source: `Splash2B:3` — Tread water 30 sec.

1. Vertical body position
2. Mouth and nose above surface
3. Sculling action of hand generates support
4. Supportive kick
5. Minimum time met

### 4. Recover object from bottom in water just over swimmer’s head

App ID: `Splash2B:4` · Short form: **Pickup just above head**

Back-page source: `Splash2B:4` — Recover object from bottom in water just over swimmer's head

1. Entire body submerged
2. Object recovered with hands and returned to surface

### 5. Wearing PFD, jump into deep water, tread 30 sec. and swim/ kick 15 m

App ID: `Splash2B:5` · Short form: **PFD jump/tread 30s/15m**

Back-page source: `Splash2B:5` — wearing PFD, jump into deep water, tread 30 sec. and swim/kick 15 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 6a. Flutter kick on Front 10 m

App ID: `Splash2B:6a` · Short form: **Front flutter 10m**

Back-page source: `Splash2B:6` — Flutter kick on Front, Back and Side (10 m)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 6b. Flutter kick on Back 10 m

App ID: `Splash2B:6b` · Short form: **Back flutter 10m**

Back-page source: `Splash2B:6` — Flutter kick on Front, Back and Side (10 m)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 6c. Flutter kick on side 10 m

App ID: `Splash2B:6c` · Short form: **Side flutter 10m**

Back-page source: `Splash2B:6` — Flutter kick on Front, Back and Side (10 m)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 7. Whip kick in vertical position 30 sec. with aid

App ID: `Splash2B:7` · Short form: **Aided vertical whip 30s**

Back-page source: `Splash2B:7` — Whip kick in vertical position 30 sec. with aid

1. Kick in vertical position
2. Kick is simultaneous and symmetrical; heels recover towards buttocks
3. Legs drive with knees apart; feet wider than knees
4. Minimum time met

### 8a. Front crawl 10 m

App ID: `Splash2B:8-front-crawl` · Short form: **Front crawl 10m**

Back-page source: `Splash2B:8` — Front crawl and Back Crawl (10 m)

1. Minimum distance completed (for both strokes)
2. **Front Crawl:** Body on front
3. **Front Crawl:** Alternate arm action
4. **Front Crawl:** Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
5. **Front Crawl:** Breathing pattern with underwater exhalation

### 8b. Back crawl 10 m

App ID: `Splash2B:8-back-crawl` · Short form: **Back crawl 10m**

Back-page source: `Splash2B:8` — Front crawl and Back Crawl (10 m)

1. Minimum distance completed (for both strokes)
2. **Back Crawl:** Body on Back
3. **Back Crawl:** Alternate arm action
4. **Back Crawl:** Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
5. **Back Crawl:** Relaxed Breathing

### 8c. Alternating front kick to side kick 10m

App ID: `Splash2B:8c` · Short form: **Front/side kick 10m**

Back-page source: `Splash2B:8-2` — Alternating front kick to side kick (10 m)

1. Completion of skills in a continuous sequence
2. Appropriate streamlined body position
3. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
4. Alternate body position every 3 sec.

### 9. Interval training: 4 x 10 m flutter kick with 20 sec. rests

App ID: `Splash2B:9` · Short form: **Flutter kick 4×10m**

Back-page source: `Splash2B:9` — Interval training 4x10m flutter kick with 20 sec. rests

1. Repetitions completed
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend

### 10a. Water Smart Messages: Swim with a Buddy

App ID: `Splash2B:10a` · Short form: **Swim with a Buddy**

Back-page source: `Splash2B:10` — Water Smart Messages:

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim with a buddy:** Never swim alone, always with a buddy

### 10b. Water Smart Messages: Wear a Lifejacket

App ID: `Splash2B:10b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash2B:10` — Water Smart Messages:

1. Participation in a water activity reinforcing the water smart messages:
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 10c. Water Smart Messages: Check the Ice

App ID: `Splash2B:10c` · Short form: **Check the Ice**

Back-page source: `Splash2B:10` — Water Smart Messages:

1. Participation in a water activity reinforcing the water smart messages:
2. **Check the Ice:** Always have an adult check ice thickness before going on it

### 10d. Water Smart Messages: Swim to Survive

App ID: `Splash2B:10d` · Short form: **Swim to Survive**

Back-page source: `Splash2B:10` — Water Smart Messages:

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim to Survive:** Follow the bubbles to the surface

## Splash 3

Source: [Splash3.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash3.html)

SHA-256: `8fcd47258c4b4624858843d7ec05cbf862a548bfac9444726c49c8e66fa7f352`

### 1. Kneeling Dive into Deep Water

App ID: `Splash3:1` · Short form: **Kneeling deep-water dive**

Back-page source: `Splash3:1` — Kneeling Dive into Deep Water

1. Head-first entry from a position on one knee with controlled return to surface

### 2. Forward Roll Entry into Deep Water

App ID: `Splash3:2` · Short form: **Deep forward roll entry**

Back-page source: `Splash3:2` — Forward Roll Entry into Deep Water

1. Controlled entry and return to surface

### 3. Tread Water 30 Seconds

App ID: `Splash3:3` · Short form: **Tread Water 30 Seconds**

Back-page source: `Splash3:3` — Tread Water 30 Seconds

1. Vertical body position
2. Mouth and nose above the surface
3. Supportive and continuous arm and leg action
4. Minimum time met

### 4. Handstand in Shallow Water

App ID: `Splash3:4` · Short form: **Shallow handstand**

Back-page source: `Splash3:4` — Handstand in Shallow Water

1. Hands on bottom; feet off bottom
2. Head fully submerged

### 5. Front Somersault (in Water)

App ID: `Splash3:5` · Short form: **Front somersault**

Back-page source: `Splash3:5` — Front Somersault (in Water)

1. Forward roll with "head over heels" rotation

### 6. Jump into deep water, tread 30 sec., and swim/kick 25 m

App ID: `Splash3:6` · Short form: **Jump/tread 30s/swim 25m**

Back-page source: `Splash3:6` — Jump into Deep Water, Tread 30 sec., and Swim/Kick 25 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 7. Flutter Kick on Back 5 m, Reverse Direction and Flutter Kick on Front 5 m

App ID: `Splash3:7` · Short form: **Back→front flutter 5m**

Back-page source: `Splash3:7` — Flutter Kick on Back 5 m; Reverse Direction and Flutter Kick on Front 5 m

1. Appropriate streamlined body position
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
3. Controlled change from back to front
4. Minimum distance completed (in both positions)

### 8. Flutter Kick on Front 5 m, Reverse Direction and Flutter Kick on Back 5 m

App ID: `Splash3:8` · Short form: **Front→back flutter 5m**

Back-page source: `Splash3:8` — Flutter Kick on Front 5 m; Reverse Direction and Flutter Kick on Back 5 m

1. Appropriate streamlined body position
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
3. Controlled change from back to front
4. Minimum distance completed (in both positions)

### 9. Whip Kick on Back 10 m

App ID: `Splash3:9` · Short form: **Whip Kick on Back 10 m**

Back-page source: `Splash3:9` — Whip Kick on Back 10 m

1. Body on back
2. Kick is simultaneous and symmetrical; heels recover towards buttocks
3. Legs drive with knees apart; feet wider than knees
4. Minimum Distance completed

### 10a. Front Crawl 15 m

App ID: `Splash3:10-front-crawl` · Short form: **Front crawl 15m**

Back-page source: `Splash3:10` — Front Crawl and Back Crawl 15 m each

1. Minimum Distance completed (for both strokes)
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
5. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 10b. Back Crawl 15 m

App ID: `Splash3:10-back-crawl` · Short form: **Back crawl 15m**

Back-page source: `Splash3:10` — Front Crawl and Back Crawl 15 m each

1. Minimum Distance completed (for both strokes)
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Hand drive finishes at thigh; recovery above surface; entry beyond shoulder
5. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
6. **Back Crawl:** Relaxed breathing

### 11. Interval Training: 4x15 m Flutter Kick with 20 Second Rests

App ID: `Splash3:11` · Short form: **Flutter kick 4×15m**

Back-page source: `Splash3:11` — Interval training: 4 x 15 m flutter kick with 20 sec. rests

1. Repetitions Completed
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed

### 12a. Water Smart Messages - Swim with a Buddy

App ID: `Splash3:12a` · Short form: **Swim with a Buddy**

Back-page source: `Splash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim with a buddy:** Always swim with someone, even in your backyard pool

### 12b. Water Smart Messages - Wear a Lifejacket

App ID: `Splash3:12b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 12c. Water Smart Messages - Check the Ice

App ID: `Splash3:12c` · Short form: **Check the Ice**

Back-page source: `Splash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Check the ice:** If the ice is safe, go with a buddy

### 12d. Water Smart Messages - Swim to Survive

App ID: `Splash3:12d` · Short form: **Swim to Survive**

Back-page source: `Splash3:12` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim to Survive:** Stay calm, take a breath, look for safety

## Splash 4

Source: [Splash4.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash4.html)

SHA-256: `7a7c3bb85b0d4fcd97c22dd206d55e41fc6742950fd71d6123bcfefe76f5d508`

### 1. Standing Dive into Deep Water

App ID: `Splash4:1` · Short form: **Standing deep-water dive**

Back-page source: `Splash4:1` — Standing Dive into Deep Water

1. Head-first entry from a standing position with controlled return to surface

### 2. Tread Water 1 minute

App ID: `Splash4:2` · Short form: **Tread Water 1 minute**

Back-page source: `Splash4:2` — Tread Water 1 minute

1. Vertical body position
2. Mouth and nose above the surface
3. Supportive and continuous arm and leg action
4. Minimum time met

### 3. Swim Underwater 5 m

App ID: `Splash4:3` · Short form: **Swim Underwater 5 m**

Back-page source: `Splash4:3` — Swim Underwater 5 m

1. Body fully submerged throughout
2. Distance completed

### 4. Canadian Swim to Survive Standard: Roll Entry into Deep Water, Tread 1 minute and Swim 50 m

App ID: `Splash4:4` · Short form: **Swim to Survive 50m**

Back-page source: `Splash4:4` — Canadian Swim to Survive Standard: Roll Entry into Deep Water, Tread 1 minute and Swim 50 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 5. Whip Kick on Front 15 m

App ID: `Splash4:5` · Short form: **Whip Kick on Front 15 m**

Back-page source: `Splash4:5` — Whip Kick on Front 15 m

1. Body on front
2. Kick is simultaneous and symmetrical; heels recover towards buttocks
3. Legs drive with knees apart; feet wider than knees
4. Minimum distance completed

### 6. Breaststroke Arms Drill 15 m

App ID: `Splash4:6` · Short form: **Breaststroke arms 15m**

Back-page source: `Splash4:6` — Breaststroke Arms Drill 15 m

1. Horizontal body position on front
2. Symmetrical arm pull towards chest; elbows higher than hands; symmetrical arm recovery beyond head
3. Regular breathing pattern with underwater exhalation
4. Minimum distance completed

### 7a. Front Crawl 25 m

App ID: `Splash4:7-front-crawl` · Short form: **Front crawl 25m**

Back-page source: `Splash4:7` — Front Crawl and Back Crawl 25 m each

1. Minimum Distance completed (for both strokes)
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
5. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 7b. Back Crawl 25 m

App ID: `Splash4:7-back-crawl` · Short form: **Back crawl 25m**

Back-page source: `Splash4:7` — Front Crawl and Back Crawl 25 m each

1. Minimum Distance completed (for both strokes)
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Hand drive finishes at thigh; recovery above surface; entry beyond shoulder
5. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
6. **Back Crawl:** Relaxed breathing

### 8. Interval Training: 4x25 m Front or Back Crawl with 20 Second Rests

App ID: `Splash4:8` · Short form: **Front/back crawl 4×25m**

Back-page source: `Splash4:8` — Interval Training: 4x25 m Front or Back Crawl with 20 Second Rests

1. Repetitions completed
2. Front crawl/back crawl must sees (See Item 7)

### 9. Sprint Front Crawl 25 m

App ID: `Splash4:9` · Short form: **Sprint Front Crawl 25 m**

Back-page source: `Splash4:9` — Sprint Front Crawl 25 m

1. Distance completed as fast as possible

### 10a. Water Smart Messages - Swim with a Buddy

App ID: `Splash4:10a` · Short form: **Swim with a Buddy**

Back-page source: `Splash4:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim with a buddy:** Always swim with someone, even in your backyard pool

### 10b. Water Smart Messages - Wear a Lifejacket

App ID: `Splash4:10b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash4:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 10c. Water Smart Messages - Check the Ice

App ID: `Splash4:10c` · Short form: **Check the Ice**

Back-page source: `Splash4:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Check the ice:** If the ice is safe, go with a buddy

### 10d. Water Smart Messages - Swim to Survive

App ID: `Splash4:10d` · Short form: **Swim to Survive**

Back-page source: `Splash4:10` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim to Survive:** Stay calm, take a breath, look for safety

## Splash 5

Source: [Splash5.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash5.html)

SHA-256: `5cd14938edac52c726e1480ad209b3c686daa534df6c01666934ceb3a9161954`

### 1. Shallow Dive into Deep Water

App ID: `Splash5:1` · Short form: **Shallow dive**

Back-page source: `Splash5:1` — Shallow Dive into Deep Water

1. Head-first entry following a path close to the surface; surface quickly

### 2. Tuck Jump (Cannonball) into Deep Water

App ID: `Splash5:2` · Short form: **Deep-water tuck jump**

Back-page source: `Splash5:2` — Tuck Jump (Cannonball) into Deep Water

1. Controlled return to surface after entry

### 3. Jump Entry into Deep Water and Tread 2 minutes

App ID: `Splash5:3` · Short form: **Deep jump & tread 2min**

Back-page source: `Splash5:3` — Jump Entry into Deep Water and Tread 2 minutes

1. Controlled return to surface after jump entry
2. Mouth and nose above surface
3. supportive arm and leg action
4. Completion of skills in a continuous sequence
5. Minimum time met

### 4. Stationary Eggbeater Kick 30 Seconds

App ID: `Splash5:4` · Short form: **Stationary eggbeater 30s**

Back-page source: `Splash5:4` — Stationary Eggbeater Kick 30 Seconds

1. Stationary kick in vertical position
2. Semi-circular leg drive toward mid-line of body; alternate drive and recovery
3. Drive with knees apart; feet wider than knees
4. Minimum time met

### 5. Back Somersault (in Water)

App ID: `Splash5:5` · Short form: **Back somersault**

Back-page source: `Splash5:5` — Back Somersault (in Water)

1. Backward roll with "head over heels" rotation

### 6. Roll Entry into Deep Water, Tread 90 Seconds and Swim 75 m

App ID: `Splash5:6` · Short form: **Roll/tread 90s/swim 75m**

Back-page source: `Splash5:6` — Roll Entry into Deep Water, Tread 90 Seconds and Swim 75 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 7. Breaststroke 25 m

App ID: `Splash5:7` · Short form: **Breaststroke 25 m**

Back-page source: `Splash5:7` — Breaststroke 25 m

1. Horizontal body position on front
2. Symmetrical arm pull towards chest; elbows higher than hands; Symmetrical arm recovery beyond head
3. Regular breathing pattern with underwater exhalation
4. Kick is simultaneous and symmetrical; heels recover towards buttocks
5. Legs drive with knees apart; feet wider than knees
6. Pull-breathe-kick-stretch sequence into full streamlined extension
7. Minimum distance completed

### 8a. Front Crawl 50 m

App ID: `Splash5:8-front-crawl` · Short form: **Front crawl 50m**

Back-page source: `Splash5:8` — Front Crawl and Back Crawl 50 m each

1. Minimum Distance completed (for both strokes)
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull below body
4. **Front Crawl:** Hand enters beyond shoulder, pull past hip
5. **Front Crawl:** Elbow higher than hand during recovery
6. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 8b. Back Crawl 50 m

App ID: `Splash5:8-back-crawl` · Short form: **Back crawl 50m**

Back-page source: `Splash5:8` — Front Crawl and Back Crawl 50 m each

1. Minimum Distance completed (for both strokes)
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Hand drive finishes at thigh; recovery above surface with elbows straight; entry beyond shoulder
5. **Back Crawl:** Drive with bent arm pull
6. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Back Crawl:** Relaxed breathing

### 9. Head-Up Front Crawl 10 m

App ID: `Splash5:9` · Short form: **Head-Up Front Crawl 10 m**

Back-page source: `Splash5:9` — Head-up Front Crawl 10 m

1. Body on front; streamlined and horizontal
2. Alternate arm action with overarm recovery and underwater pull below body; hand enters beyond shoulder; pull past hip
3. Elbow higher than hand during recovery
4. Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
5. Eyes above surface
6. Minimum distance completed

### 10. Interval Training: 4x50 m Front or Back Crawl with 30 sec. Rests

App ID: `Splash5:10` · Short form: **Front/back crawl 4×50m**

Back-page source: `Splash5:10` — Interval Training:4x50 m Front or Back Crawl with 30 sec Rests

1. Repetitions completed
2. Front crawl/back crawl must sees (see item 8)

### 11. Interval Training: 4x15 m Breaststroke with 30 second Rests

App ID: `Splash5:11` · Short form: **Breaststroke 4×15m**

Back-page source: `Splash5:11` — Interval Training: 4x15 m Breaststroke with 30 sec Rests

1. Repetitions completed
2. Breaststroke must sees (see item 7)

### 12a. Sprint Front Crawl 25 m each

App ID: `Splash5:12a` · Short form: **Front crawl sprint 25m**

Back-page source: `Splash5:12` — Sprint Front Crawl and Back Crawl 25 m each

1. Distance completed as fast as possible

### 12b. Sprint Back Crawl 25 m each

App ID: `Splash5:12b` · Short form: **Back crawl sprint 25m**

Back-page source: `Splash5:12` — Sprint Front Crawl and Back Crawl 25 m each

1. Distance completed as fast as possible

### 13a. Water Smart Messages - Swim with a Buddy

App ID: `Splash5:13a` · Short form: **Swim with a Buddy**

Back-page source: `Splash5:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim with a buddy:** The safest place to play and swim is in an area supervised by Lifeguards

### 13b. Water Smart Messages - Wear a Lifejacket

App ID: `Splash5:13b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash5:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 13c. Water Smart Messages - Check the Ice

App ID: `Splash5:13c` · Short form: **Check the Ice**

Back-page source: `Splash5:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Check the ice:** If you fall through the ice, don't panic. Kick your feet, push yourself onto the unbroken ice on your stomach, roll away from the broken ice.

### 13d. Water Smart Messages - Swim to Survive

App ID: `Splash5:13d` · Short form: **Swim to Survive**

Back-page source: `Splash5:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim to Survive:** Swim, rest, swim, get there.

## Splash 6

Source: [Splash6.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash6.html)

SHA-256: `f324b967060f8b0b01defa8721eceb7e47b4aeec217d0ab9c4510da2ff07175c`

### 1. Stride Entry into Deep Water

App ID: `Splash6:1` · Short form: **Deep-water stride entry**

Back-page source: `Splash6:1` — Stride Entry into Deep Water

1. Foot-first entry with head above surface

### 2. Compact Jump into Deep Water

App ID: `Splash6:2` · Short form: **Deep-water compact jump**

Back-page source: `Splash6:2` — Compact Jump into Deep Water

1. Vertical, foot-first entry
2. Crossed legs, arms tight to body
3. Airway protected

### 3. Legs-Only Surface Support 45 Seconds

App ID: `Splash6:3` · Short form: **Legs-only support 45s**

Back-page source: `Splash6:3` — Legs-only Surface Support 45 Seconds

1. Vertical body position
2. Mouth and nose above surface
3. Recognizable supportive lifesaving kick
4. Minimum time met

### 4. Swim Underwater 10 m to recover object

App ID: `Splash6:4` · Short form: **Underwater recovery 10m**

Back-page source: `Splash6:4` — Swim Underwater 10 m to Recover Object

1. Body fully submerged throughout
2. Minimum distance completed
3. Object recovered and returned to surface

### 5. Eggbeater Kick on Back 15 m

App ID: `Splash6:5` · Short form: **Back eggbeater 15m**

Back-page source: `Splash6:5` — Eggbeater Kick on Back 15 m

1. Body on back
2. Semi-circular leg drive toward mid-line of body; alternate drive and recovery
3. Drive with knees apart; feet wider than knees
4. Minimum distance completed

### 6. Scissor Kick 15 m

App ID: `Splash6:6` · Short form: **Scissor Kick 15 m**

Back-page source: `Splash6:6` — Scissor Kick 15 m

1. Body on side
2. Both legs kick at same time nd recover at the same time
3. One leg extends forward and other leg extends back to begin drive
4. Legs squeeze together in scissor-like action for drive phase
5. Minimum distance completed

### 7. Breaststroke 50 m

App ID: `Splash6:7` · Short form: **Breaststroke 50 m**

Back-page source: `Splash6:7` — Breaststroke 50 m

1. Horizontal body position on front
2. Symmetrical arm pull towards chest; elbows higher than hands; Symmetrical arm recovery beyond head
3. Regular breathing pattern with underwater exhalation
4. Kick is simultaneous and symmetrical; heels recover towards buttocks
5. Legs drive with knees apart; feet wider than knees
6. Pull-breathe-kick-stretch sequence into full streamlined extension
7. Minimum distance completed

### 8a. Front Crawl 100 m

App ID: `Splash6:8-front-crawl` · Short form: **Front crawl 100m**

Back-page source: `Splash6:8` — Front Crawl and Back Crawl 100 m each

1. Minimum Distance completed (for both strokes)
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull below body
4. **Front Crawl:** Hand enters beyond shoulder, pull past hip
5. **Front Crawl:** Elbow higher than hand during recovery
6. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 8b. Back Crawl 100 m

App ID: `Splash6:8-back-crawl` · Short form: **Back crawl 100m**

Back-page source: `Splash6:8` — Front Crawl and Back Crawl 100 m each

1. Minimum Distance completed (for both strokes)
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Hand drive finishes at thigh; recovery above surface with elbows straight; entry beyond shoulder
5. **Back Crawl:** Drive with bent arm pull
6. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Back Crawl:** Relaxed breathing

### 9. Head-up Swim 25 m

App ID: `Splash6:9` · Short form: **Head-up Swim 25 m**

Back-page source: `Splash6:9` — Head-up Swim 25 m

1. Eyes above surface
2. Minimum distance completed

### 10. Interval Training: 4x25 m Breaststroke with 30 sec. Rests

App ID: `Splash6:10` · Short form: **Breaststroke 4×25m**

Back-page source: `Splash6:10` — Interval Training: 4x25 m Breaststroke with 30 second Rests

1. Repetitions completed
2. Breaststroke must sees (see item 7)

### 11. Sprint Breaststroke 25 m

App ID: `Splash6:11` · Short form: **Breaststroke sprint 25m**

Back-page source: `Splash6:11` — Sprint Breaststroke 25 m each

1. Distance completed as fast as possible

### 12. Workout 300 m 15 sec. rests; 4 x 25m back crawl with 15 sec. rests;

App ID: `Splash6:12` · Short form: **Workout 300m**

Back-page source: `Splash6:12` — Workout 300 m; 50 m Warm-up (Choice of Strokes); 4x25 m Front Crawl with 15 Sec Rests; 4x25 m Back Crawl with 15 Sec Rests; 50 m Cool-Down (Choice of Strokes)

1. Workout Completed

### 13a. Water Smart Messages - Swim with a Buddy

App ID: `Splash6:13a` · Short form: **Swim with a Buddy**

Back-page source: `Splash6:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim with a buddy:** The safest place to play and swim is in an area supervised by lifeguards

### 13b. Water Smart Messages - Wear a Lifejacket

App ID: `Splash6:13b` · Short form: **Wear a Lifejacket**

Back-page source: `Splash6:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Wear a Lifejacket:** Is everyone in my family wearing lifejackets in the boat including me? Mom? Dad? Others?

### 13c. Water Smart Messages - Check the ice

App ID: `Splash6:13c` · Short form: **Check the Ice**

Back-page source: `Splash6:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Check the ice:** If you fall through the ice, don't panic, kick your feet, push yourself onto the unbroken ice on your stomach, roll away from the broken ice

### 13d. Water Smart Messages - Swim to Survive

App ID: `Splash6:13d` · Short form: **Swim to Survive**

Back-page source: `Splash6:13` — Water Smart Messages

1. Participation in a water activity reinforcing the Water Smart messages:
2. **Swim to Survive:** Swim, rest, swim, get there

## Splash 7

Source: [Splash7.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash7.html)

SHA-256: `c2f0b272086888723af0f52949c2ac5a2458812ca816cc9ed75dad15e7740b1c`

### 1. Head-up Front Crawl or Breaststroke 25 m. or yd.

App ID: `Splash7:1` · Short form: **Head-up swim 25m/yd**

Back-page source: `Splash7:1` — Swim 25 m or yd Head-up Front Crawl or Breaststroke

1. Basic arm and leg mechanics
2. Distance completed non-stop
3. Eyes above the surface

### 2. Scull in ready position - 30 Sec.

App ID: `Splash7:2` · Short form: **Ready position/scull 30s**

Back-page source: `Splash7:2` — Demonstrate Ready Position with a Stationary Scull for 30 Sec.

1. Ready position maintained with toes at or near surface
2. Effective support scull showing stationary control
3. Time met or bettered

### 3. Carry object 2.3 Kg (5 lb) 15 m or yd.

App ID: `Splash7:3` · Short form: **Carry 2.3kg for 15m/yd**

Back-page source: `Splash7:3` — Carry a 2.3 kg(5 lb) Object 15 m or yd Using Swimmer's Choice of Lifesaving Kick

1. Basic lifesaving kick mechanics with control of direction
2. Distance completed non-stop
3. Head remains above the surface

### 4a. Foot-first Surface Dive

App ID: `Splash7:4a` · Short form: **Foot-first Surface Dive**

Back-page source: `Splash7:4` — Demonstrate a Foot-First and a Head-First Surface Dive to a Maximum Depth of 2 m

1. Vertical descent
2. Depth attained with ease
3. Arm extended above the head on ascent

### 4b. Head-first Surface Dive

App ID: `Splash7:4b` · Short form: **Head-first Surface Dive**

Back-page source: `Splash7:4` — Demonstrate a Foot-First and a Head-First Surface Dive to a Maximum Depth of 2 m

1. Vertical descent
2. Depth attained with ease
3. Arm extended above the head on ascent

### 5. Lifesaving Kick - 25 m. or yd.

App ID: `Splash7:5` · Short form: **Legs-only swim 25m/yd**

Back-page source: `Splash7:5` — Demonstrate 25 m or yd Legs-Only Using Swimmer's Choice of Lifesaving Kick

1. Basic lifesaving kick mechanics with control of direction
2. Distance completed non-stop

### 6. Inflate Clothing /Use as a Buoyant Assist

App ID: `Splash7:6` · Short form: **Inflate 2 clothing items**

Back-page source: `Splash7:6` — Demonstrate the Ability to Inflate 2 Items of Clothing and Use as a Buoyant Assist

1. Clothing items inflated
2. Candidate Floats Supported by the inflated clothing item

### 7. Obstacle Swim - 25 m. or yd.

App ID: `Splash7:7` · Short form: **Obstacle swim 25m/yd**

Back-page source: `Splash7:7` — Swim 25 m or yd Using Any Stroke. Demonstrate the Ability to Swim Under an Obstacle Located at the Halfway Point

1. Candidate swims under the obstacle
2. Distance completed

### 8a. Front Crawl - 50 m or yd

App ID: `Splash7:8-front-crawl` · Short form: **Front crawl 50m/yd**

Back-page source: `Splash7:8` — Swim Each of the Following Strokes 50 m or yd Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Front Crawl:** Hand entry beyond shoulder; pull past hip
4. **Front Crawl:** Alternate leg action drive
5. **Front Crawl:** Rhythmic kick
6. **Front Crawl:** Regular breathing pattern to side with underwater exhalation
7. **Front Crawl:** Continuous propulsion

### 8b. Back Crawl - 50 m or yd

App ID: `Splash7:8-back-crawl` · Short form: **Back crawl 50m/yd**

Back-page source: `Splash7:8` — Swim Each of the Following Strokes 50 m or yd Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Back Crawl:** Hand entry beyond shoulder; pull past hip
4. **Back Crawl:** Alternate leg action drive
5. **Back Crawl:** Rhythmic kick
6. **Back Crawl:** Continuous propulsion

### 8c. Breaststroke - 50 m or yd

App ID: `Splash7:8-breaststroke` · Short form: **Breaststroke 50m/yd**

Back-page source: `Splash7:8` — Swim Each of the Following Strokes 50 m or yd Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Breaststroke:** Symmetrical arm recovery and drive
3. **Breaststroke:** Regular breathing pattern with underwater exhalation
4. **Breaststroke:** Both legs kick at same time and recover at the same time
5. **Breaststroke:** Drive with knees apart; feet wider than knees

### 9. Fitness training - 350 m. or yd. workout (3 times)

App ID: `Splash7:9` · Short form: **Workout 350m/yd ×3**

Back-page source: `Splash7:9` — Complete a 350 m or yd workout on 3 different Occasions Throughout the Training Course

1. **Warm-up::** 100 m or yd Swimmer's choice
2. **Work Set::** 6 x 25 m or yd any Stroke on 60 sec
3. **Cool Down::** 100 m or yd Swimmer's choice
4. **Cool Down::** Workout is completed at least 3 times during the training course
5. **Cool Down::** Ability to locate pulse and count beats per minute

### 10. Swim 100 m in 3 min (100 yd in 2:40 min) or better

App ID: `Splash7:10` · Short form: **100m 3min / 100yd 2:40**

Back-page source: `Splash7:10` — Swim 100 m in 3 min. or Better (100 yd in 2:40 min) Using any Stroke or Combination of Strokes of the Swimmer's Choice

1. Recognizable stroke(s) used
2. Distance completed non-stop
3. Time met or bettered

### 11. Assist conscious victim

App ID: `Splash7:11` · Short form: **Conscious assessment**

Back-page source: `Splash7:11` — Demonstrate primary assessment of a conscious and cooperative victim who describes his or her chief complaint and how injury occurred

1. Assessment of environment for hazards
2. Establish responsiveness
3. Check for obvious signs of bleeding
4. Collection of information from victim about chief complaint and how injury occurred
5. Activate EMS (if required)

### 12. Contacting EMS

App ID: `Splash7:12` · Short form: **Calling EMS**

Back-page source: `Splash7:12` — Demonstrate the Ability to Recognize When to Call EMS and How To Do It

1. Basic understanding demonstrated through performance
2. Ability to give examples of circumstances requiring an EMS call
3. Ability to demonstrate an EMS call providing the 4 Ws (who, why, what where)

### 13. Care for External Bleeding

App ID: `Splash7:13` · Short form: **External bleeding care**

Back-page source: `Splash7:13` — Demonstrate Primary Assessment and Appropriate Care for External Bleeding

1. Assessment of environment for hazards
2. Establish responsiveness
3. Assess ABCs
4. Check for obvious signs of bleeding
5. Direct Pressure
6. Elevation if appropriate
7. Rest and reassurance for victim
8. Recruit bystanders to assist rescuer
9. Activate EMS (if required)

### 14. Look and See

App ID: `Splash7:14` · Short form: **Observe & recall 10–15s**

Back-page source: `Splash7:14` — Look at the Aquatic Facility for 10-15 Sec. Turn and Describe What You Saw

1. Recognition of (any) seven elements in the aquatic environment
2. Ability to identify at least one element that might be a "hot spot" for potential trouble

### 15a. Victim simulate: a Weak Swimmer and a Non-Swimmer

App ID: `Splash7:15a` · Short form: **Act weak/non-swim**

Back-page source: `Splash7:15a` — Demonstrate Ability to Simulate in the Water the Appearance of a Weak Swimmer and a Non-Swimmer

1. Behaviour differentiating between the appearance of a weak swimmer and a non-swimmer

### 15b. Victim recognition: a Weak Swimmer and Non-Swimmer

App ID: `Splash7:15b` · Short form: **Spot weak/non-swim**

Back-page source: `Splash7:15b` — Demonstrate Ability to Recognize the Difference Between a Weak Swimmer and Non-Swimmer

1. Ability to identify the difference between the appearance of a weak swimmer and a non-swimmer
2. Understanding of potential risk to rescuer safety from these two victim types

### 16. Rescue drill: throw aid to target - 30 Sec.

App ID: `Splash7:16` · Short form: **Throw aid with line 30s**

Back-page source: `Splash7:16` — From a standing Position on Land, Throw a Buoyant Aid With Line to Hit a Target on the Surface of the Water as Many Times as Possible Within 30 sec.

1. At least 3 "hits" within time limit

## Splash 8

Source: [Splash8.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash8.html)

SHA-256: `1adc2a051bcbd2f6f79611df21f933459c770cbad8d01d7f2831e0558c093b02`

### 1. Somersault sequence - forward & backward

App ID: `Splash8:1` · Short form: **Front/back somersaults**

Back-page source: `Splash8:1` — Demonstrate 1 Forward And 1 Backward Somersault in the Water as a Continuous Sequence

1. Complete "head over heels rotation" with control of direction
2. Somersaults performed as a continuous sequence

### 2. Stride entry & swim head-up - 25 m. or yr. Ready position

App ID: `Splash8:2` · Short form: **Stride/head-up 25m/yd**

Back-page source: `Splash8:2` — Stride Entry and Swim 25 m or yd Head-up Front Crawl or Breaststroke. Assume Ready Position

1. Eyes above surface
2. Basic arm and leg mechanics
3. Distance completed non-stop
4. Ready position

### 3. Eggbeater kick on back - 25 m. or yd.

App ID: `Splash8:3` · Short form: **Back eggbeater 25m/yd**

Back-page source: `Splash8:3` — Demonstrate 25 m or yd Eggbeater Kick on Back

1. Movement pattern of kick resembles whip kick with alternate leg action
2. Drive with knees apart; feet wider than knees
3. Distance completed

### 4. Support object (2.3 kg.) - 1 min.

App ID: `Splash8:4` · Short form: **Support 2.3kg for 1min**

Back-page source: `Splash8:4` — Support a 2.3 kg (5 lb) Object for at Least 1 min. in Deep Water

1. Head above surface throughout
2. Time met or bettered

### 5. Search and recover object

App ID: `Splash8:5` · Short form: **Dive 2m/recover 2–3m/yd**

Back-page source: `Splash8:5` — Head-up Approach into Surface Dive to a Maximum Depth 2 m Swim Underwater for At Least 2-3 m or yd to Recover a Small Object. Surface and Carry Object to Starting Point

1. Eyes up and forward on approach
2. Vertical descent
3. Travel in designated direction
4. Eyes open underwater
5. Object recovered and returned to starting point
6. Arm extended above head on ascent

### 6. Demonstrate an Assisted Removal of a Conscious Victim

App ID: `Splash8:6` · Short form: **Conscious victim removal**

Back-page source: `Splash8:6` — Demonstrate an Assisted Removal of a Conscious Victim

1. Victim removed safely with concern for victim comfort
2. Appropriate choice of removal
3. Prevention of injury to rescuer and victim

### 7a. Front Crawl - 75 m. or yd.

App ID: `Splash8:7-front-crawl` · Short form: **Front crawl 75m/yd**

Back-page source: `Splash8:7` — Swim Each of the Following Strokes 75 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum distance completed (for each stroke)
2. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Front Crawl:** Hand entry beyond shoulder; pull past hip
4. **Front Crawl:** Alternate leg action drive
5. **Front Crawl:** Rhythmic kick
6. **Front Crawl:** Regular breathing pattern to side with underwater exhalation
7. **Front Crawl:** Continuous propulsion

### 7b. Back Crawl - 75 m. or yd.

App ID: `Splash8:7-back-crawl` · Short form: **Back crawl 75m/yd**

Back-page source: `Splash8:7` — Swim Each of the Following Strokes 75 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum distance completed (for each stroke)
2. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Back Crawl:** Hand entry beyond shoulder; pull past hip
4. **Back Crawl:** Alternate leg action drive
5. **Back Crawl:** Rhythmic kick
6. **Back Crawl:** Continuous propulsion

### 7c. Breaststroke - 75 m. or yd.

App ID: `Splash8:7-breaststroke` · Short form: **Breaststroke 75m/yd**

Back-page source: `Splash8:7` — Swim Each of the Following Strokes 75 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum distance completed (for each stroke)
2. **Breaststroke:** Symmetrical arm recovery and drive
3. **Breaststroke:** Regular breathing pattern with underwater exhalation
4. **Breaststroke:** Both legs kick at same time and recover at the same time
5. **Breaststroke:** Drive with knees apart; feet wider than knees
6. **Breaststroke:** Continuous propulsion

### 8. Lifesaving medley - 100 M. or yd. (3 min)

App ID: `Splash8:8` · Short form: **Medley 100m/yd 3min**

Back-page source: `Splash8:8` — Swim 100 m or yd Individual Medley as Follows: 25 m or yd Each of Lifesaving Kick, Back Crawl, Breaststroke, Front Crawl

1. Non-stop medley completed at least 3 times during training course
2. Ability to locate pulse and count beats per minute

### 9. Swim 200 m in 6 min (200 yd in 5:20 min) or better

App ID: `Splash8:9` · Short form: **200m 6min / 200yd 5:20**

Back-page source: `Splash8:9` — Swim 200 m in 6 min. or Better (200 yd in 5:20 min) Using Any Stroke or Combination of Strokes of the Swimmer's Choice

1. Recognizable stroke(s) used
2. Distance completed non-stop
3. Time met or bettered

### 10. Assess unconscious, breathing Victim

App ID: `Splash8:10` · Short form: **Unconscious assessment**

Back-page source: `Splash8:10` — Demonstrate a Primary Assessment Including Hazards and ABCs on an Unconscious, Breathing Victim

1. Assessment of environment for hazards
2. Establish unresponsiveness
3. Activate emergency medical system
4. Position victim (turn if necessary)
5. Open airway
6. Check for breathing; look, listen and feel (no more than 5 sec)
7. Semi-prone position

### 11. Care for a Victim in Shock

App ID: `Splash8:11` · Short form: **Shock care**

Back-page source: `Splash8:11` — Demonstrate Emergency Care for a Victim in Shock

1. Assessment of environment for hazards
2. Establish responsiveness
3. Assess ABCs
4. Activate EMS (if required)
5. Maintenance of natural warmth
6. Rest and reassurance
7. Semi-prone position (unless injury dictates otherwise)
8. Bystanders recruited to assist rescuer

### 12a. Obstructed Airway - conscious victim

App ID: `Splash8:12a` · Short form: **Obstructed airway care**

Back-page source: `Splash8:12a` — Simulate the Appearance and Treatment of a Conscious Adult or Child Victim with an Obstructed Airway

1. **Assessment of degree of obstruction:** ask: "Are you choking?"
2. **Assessment of degree of obstruction:** Selection of appropriate procedures:

### 12b. Obstructed Airway - Mild obstruction victim

App ID: `Splash8:12b` · Short form: **Mild airway obstruction**

Back-page source: `Splash8:12b` — Simulate the Appearance and Treatment of a Conscious Adult or Child Victim with a Mild Obstructed Airway

1. Coughing encouraged
2. Reassurance for victim

### 12c. Obstructed Airway - Severe obstruction victim

App ID: `Splash8:12c` · Short form: **Severe airway blockage**

Back-page source: `Splash8:12c` — Simulate the Appearance and Treatment of a Conscious Adult or Child Victim with a Severe Obstructed Airway

1. Shout for help
2. Careful land marking
3. Repeated abdominal thrusts (chest thrusts for Pregnant or obese victim) until airway is clear
4. If successful, victim directed to see a physician to rule out complications from the obstruction or treatment

### 13a. Victim Simulation: weak Swimmer, a non-Swimmer and unconscious

App ID: `Splash8:13a` · Short form: **Simulate victim types**

Back-page source: `Splash8:13a` — Demonstrate Ability to Simulate in the Water the Appearance of a Weak Swimmer, a Non-swimmer and an Unconscious Victim

1. Behaviour differentiating between the appearance of a weak swimmer an a non-swimmer; and the behaviour consistent with that of an unconscious victim

### 13b. Victim Recognition: weak swimmer and non-swimmer, unconscious Victim

App ID: `Splash8:13b` · Short form: **Identify victim types**

Back-page source: `Splash8:13b` — Demonstrate the Ability to Recognize the Difference Between a Weak Swimmer a Non-swimmer and an Unconscious Victim

1. Ability to identify the difference between the appearance of a weak swimmer and a non-swimmer; and ability to recognize the unconscious victim

### 13c. Victim Avoidance

App ID: `Splash8:13c` · Short form: **Avoid dangerous victims**

Back-page source: `Splash8:13c` — Demonstrate the Ability to recognize and Avoid Victims Who Present a Danger to the Rescuer

1. Recognition of risk or threat to personal safety
2. Speed of avoidance response including use of ready position or back/swim away or submerging quickly
3. Calm, clear direction to victim

### 14. Rescue with buoyant aid - 20 m. or yd.

App ID: `Splash8:14` · Short form: **No-contact rescue 20m/yd**

Back-page source: `Splash8:14` — Perform a Non-Contact Rescue Using a Buoyant Aid. Approach 20 m or yd and Encourage Victim to Safety While Maintaining a Safe Distance and Calling for Assistance. Rescuer Performs Appropriate Follow up Procedures, Including Treatment for Shock

1. Quick accurate recognition
2. Personal safety maintained
3. Appropriate assessment of situation
4. Call for help
5. Appropriate choice and use of aid
6. Appropriate entry and efficient head-up approach
7. Ready position at a safe distance
8. Provision of aid to victim at earliest opportunity
9. Victim secured at nearest point of safety
10. Appropriate care of victim and follow-up
11. Visual contact and constructive communication
12. Victim reassured until rescuer relieved of responsibility

## Splash 9

Source: [Splash9.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/Splash9.html)

SHA-256: `e1afd5108567b56fa74d17a7437f536095c8a038ac448031719cb7cc760c8e78`

### 1. Entries with aids - at least 2

App ID: `Splash9:1` · Short form: **Two entries with aids**

Back-page source: `Splash9:1` — Demonstrate At Least 2 Different Entries with Different Aids

1. Eyes up and forward on entry
2. Control of aid on entry

### 2. Head-up swim (25 m. or yd.) & scull in ready position.

App ID: `Splash9:2` · Short form: **Aid/head-up swim 25m/yd**

Back-page source: `Splash9:2` — Entry With Aid and Swim 25 m or yd Head-up Front Crawl or Breaststroke. Assume Ready Position and Demonstrate Ability to Scull Forward, Backward and Turn

1. Control of aid on entry
2. Eyes above the surface during swim
3. Basic arm and leg mechanics
4. Distance completed non-stop
5. Ready position with effective support scull showing ability to move forward, backward and turn

### 3. Defence Methods - Front, Side and Rear

App ID: `Splash9:3` · Short form: **Front/side/rear defence**

Back-page source: `Splash9:3` — Demonstrate Defence Methods from the Front, Side and Rear

1. Avoidance of grasp(s)
2. Ready position at a safe distance
3. Constructive communication
4. Treading water throughout

### 4. Eggbeater Kick - travel, change direction & height

App ID: `Splash9:4` · Short form: **Eggbeater travel & turns**

Back-page source: `Splash9:4` — Demonstrate Eggbeater Kick Showing Ability to Travel, Change Direction and Height Levels

1. Vertical body position
2. Movement pattern of kick resembles whip kick with alternate leg action
3. Drive with knees apart; feet wider than knees
4. Head above surface throughout; change in height on command
5. Travel 1-2 m or yd in any direction

### 5. Carry object 4.5 kg (10 lb) Object - 25 m. or yd.

App ID: `Splash9:5` · Short form: **Carry 4.5kg for 25m/yd**

Back-page source: `Splash9:5` — Carry a 4.5 kg (10 lb) Object 25 m or yd Using Swimmer's Choice of Lifesaving Kick

1. Basic lifesaving kick mechanics with control of direction
2. Distance completed non-stop
3. Head remains above the surface

### 6. Remove unconscious victim

App ID: `Splash9:6` · Short form: **Unconscious removal**

Back-page source: `Splash9:6` — Demonstrate the Removal of an Unconscious Victim with the Assistance of an Untrained Bystander

1. Clear and appropriate direction for bystander
2. Appropriate choice of removal
3. Safe removal without injury to rescuer, bystander or victim

### 7. Search to recover object

App ID: `Splash9:7` · Short form: **Dives/recovery 5–10m/yd**

Back-page source: `Splash9:7` — Head-up Approach into Head-first Surface Dive to a Maximum Depth of 2 m Swim Underwater for 5-10 m or yd and Surface. Foot-first Surface Dive (max depth 2m) Recover an Object and Return It to the Starting Point

1. Eyes up and forward on approach
2. Vertical descent
3. Travel in designated direction
4. Eyes open underwater
5. Object recovered and returned
6. Arm extended above head on ascent

### 8. Turn & support victim face-up - shallow water

App ID: `Splash9:8` · Short form: **Turn & support face-up**

Back-page source: `Splash9:8` — Demonstrate in Shallow Water, the Ability to Turn a Victim Face-up and Support the Face Above the Surface

1. Victim's face supported above surface
2. Airway maintained throughout; water prevented from entering victim's nose or mouth
3. Victim quickly secured at nearest point of safety

### 9. Front Crawl, 100 m or y.

App ID: `Splash9:9-front-crawl` · Short form: **Front crawl 100m/yd**

Back-page source: `Splash9:9` — Swim Each of the Following Strokes 100 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Front Crawl:** Hand entry beyond shoulder, pull past hip
4. **Front Crawl:** Alternate leg action drive
5. **Front Crawl:** Rhythmic kick
6. **Front Crawl:** Regular breathing pattern to side with underwater exhalation
7. **Front Crawl:** Continuous propulsion

### 9. Back Crawl, 100 m or y.

App ID: `Splash9:9-back-crawl` · Short form: **Back crawl 100m/yd**

Back-page source: `Splash9:9` — Swim Each of the Following Strokes 100 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
3. **Back Crawl:** Hand entry beyond shoulder, pull past hip
4. **Back Crawl:** Alternate leg action drive
5. **Back Crawl:** Rhythmic kick
6. **Back Crawl:** Continuous propulsion

### 9. Breaststroke, 100 m or y.

App ID: `Splash9:9-breaststroke` · Short form: **Breaststroke 100m/yd**

Back-page source: `Splash9:9` — Swim Each of the Following Strokes 100 m or yd: Front Crawl, Back Crawl, Breaststroke

1. Minimum Distance completed (for each stroke)
2. **Breaststroke:** Symmetrical arm recovery and drive
3. **Breaststroke:** Regular breathing pattern with underwater exhalation
4. **Breaststroke:** Both legs kick at same time and recover at the same time
5. **Breaststroke:** Drive with knees apart; feet wider than knees
6. **Breaststroke:** Continuous propulsion

### 10. Fitness training - 600 m. or yd. workout (3 times)

App ID: `Splash9:10` · Short form: **Workout 600m/yd ×3**

Back-page source: `Splash9:10` — Complete a 600 m or yd Workout At Least 3 Times During the Training Course.

1. **Warm-up::** 25 m or yd Lifesaving kick,
2. **Warm-up::** 50 m or yd Back Crawl,
3. **Warm-up::** 50 m or yd Breaststroke,
4. **Warm-up::** 50 m or yd Front Crawl
5. **Stroke drills::** 5 x 25 m or yd Instructor's choice
6. **Work Set::** 4 x 50 m or yd on 90 sec Instructor's choice
7. **Cool down::** 100 m or yd Swimmer's choice
8. Workout is completed at least 3 times during the training course
9. Ability to locate pulse and count beats per minute
10. Ability to adapt pace in accordance with desired heart rate target

### 11. Swim 300 m in 9 min (300 yd in 8:00 min) or better

App ID: `Splash9:11` · Short form: **300m 9min / 300yd 8min**

Back-page source: `Splash9:11` — Swim 300 m in 9 min. or Better (300 yd in 8:00 min) Using Any Stroke or Combination of Strokes of the Swimmer's Choice

1. Recognizable stroke(s) used
2. Distance completed with a uniform pace
3. Time limit met or bettered

### 12. Care of a Bone or Joint Injury

App ID: `Splash9:12` · Short form: **Bone/joint injury care**

Back-page source: `Splash9:12` — Demonstrate the Recognition and Care of a Bone or Joint Injury

1. Assessment of environment for hazards
2. Establish responsiveness
3. Recognition of injury based on pain, obvious deformity or swelling and the probable cause
4. Movement of injury minimized to prevent further injury
5. If available, ice applied to injury site
6. EMS activated of necessary

### 13. Care for respiratory emergency - asthma or allergic reaction

App ID: `Splash9:13` · Short form: **Asthma/allergy care**

Back-page source: `Splash9:13` — Demonstrate the Recognition and Care of a Victim Suffering Respiratory Distress from Asthma or Severe Allergic Reaction

1. Assessment of environment for hazards
2. Establish responsiveness
3. Determination of cause (and removal if possible) of emergency
4. Victim placed in most comfortable position
5. Victim assisted to find and use medication (asthma inhaler or epinephrine auto-injector)
6. Activate EMS if necessary

### 14. Locate and describe submerged object

App ID: `Splash9:14` · Short form: **Locate underwater object**

Back-page source: `Splash9:14` — From 3 Different Heights or Positions, Locate and Describe an Object Located on the Bottom or Below the Surface

1. Ability to identify the barriers to recognition
2. Move to a position to enable candidate to best "see" the object

### 15. Rescue with towing aid - 20 m. or yd.

App ID: `Splash9:15` · Short form: **Towing-aid rescue 20m/yd**

Back-page source: `Splash9:15` — Perform a Rescue of a Weak or Non-Swimmer with a Towing Aid. Swim Approach 20 m or yd and Tow Victim to Safety Showing Ability to Avoid Contact. Rescuer Performs Appropriate Follow Up Procedures, Including Treatment for Shock

1. Quick accurate recognition
2. Personal safety maintained
3. Appropriate assessment of situation
4. Call for help
5. Appropriate choice and use of aid
6. Appropriate entry and efficient approach
7. Appropriate choice of tow for the circumstance
8. Victim secured at nearest point of safety
9. Appropriate care of victim and follow-up
10. Visual contact and constructive communication throughout
11. Appropriate use of bystanders when available
12. Victim observed/reassured until rescuer relieved of responsibility

## Splash Fitness

Source: [SplashFitness.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/SplashFitness.html)

SHA-256: `795701141af1a59c163846b785710f3b989f2a3e00cba522e655206b0608ae79`

### 1. Pace Clocks and Timers

App ID: `SplashFitness:1` · Short form: **Pace Clocks and Timers**

Back-page source: `SplashFitness:1` — Pace Clocks and Timers

1. Use of a timer to measure swimming time over a specific distance
2. Use of a timer to measure and control rest time between intervals
3. Use of a timer to measure personal pulse rate during training

### 2. Stretches for Swimmers

App ID: `SplashFitness:2` · Short form: **Stretches for Swimmers**

Back-page source: `SplashFitness:2` — Stretches for Swimmers

1. Four Stretches for swimmers demonstrated

### 3. Kicking Interval Training: 4 x 25 m Flutter Kick and/or Whip Kick with 15-20 sec Rests

App ID: `SplashFitness:3` · Short form: **Flutter/whip kick 4×25m**

Back-page source: `SplashFitness:3` — Kicking Interval Training: 4 x 25 m Flutter Kick and /or Whip Kick with 15-20 sec. Rests

1. Repetitions completed
2. Propulsive flutter kick with alternate leg action drive and rhythmic kick and/or whip kick with simultaneous kick and recovery and drive with knees apart; feet wider than knees

### 4. Swimming Interval Training: Swim Interval Sets Selected by the Swimmer

App ID: `SplashFitness:4` · Short form: **Swim interval sets**

Back-page source: `SplashFitness:4` — Swimming Interval Training: Swim Interval Sets Selected by the Swimmer

1. Repetitions completed
2. Recognizable stroke(s) used

### 5. Workout 300 m.

App ID: `SplashFitness:5` · Short form: **Workout 300m**

Back-page source: `SplashFitness:5` — Workout (300 m): 50 m Warm-up, 4 x 25 m Front or Back Crawl with 15 sec. rests, 4 x 25 m Choice of Strokes with 15 sec. rests, 50 m Cool-Down

1. **Warm-up::** 50 m (choice of Strokes)
2. **main set(s):** 4 x 25 m front or back crawl with 15 sec rests
3. **main set(s):** 4 x 25 m (choice of strokes) with 15 sec rests
4. **Cool Down::** 50 m (choice of strokes)
5. Workout completed

### 6. Workout design: Design and Demonstrate 2 Sample Workouts

App ID: `SplashFitness:6` · Short form: **Design 2 workouts**

Back-page source: `SplashFitness:6` — Workout Design: Design and Demonstrate 2 Sample Workouts

1. Workout includes warm-up, main set(s) and cool-down
2. Workouts completed

### 7. Distance Swim: Endurance Challenge

App ID: `SplashFitness:7` · Short form: **Distance/endurance swim**

Back-page source: `SplashFitness:7` — Distance Swim: Endurance Challenge

1. Continuous swim without stopping or resting
2. Distance completed

### 8. Sprint Swim: 25 m each for at Least 2 Strokes Chosen by the Swimmer

App ID: `SplashFitness:8` · Short form: **Two-stroke sprint 25m**

Back-page source: `SplashFitness:8` — Sprint Swim: 25 m each for at least 2 Strokes Chosen by the Swimmer

1. Distance completed
2. Measure personal pulse rate after each swim

## Splash Private

Source: [SplashPrivate.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/SplashPrivate.html)

SHA-256: `6c6bb454b867c8465b4691d3e0c646a06b6552205b3f3454cd7a9630c207898e`

Reference levels: Splash 1, Splash 2, Splash 3, Splash 4, Splash 5, Splash 6, Splash 7, Splash 8, Splash 9, Teen/Adult 1, Teen/Adult 2, Teen/Adult 3, Little Splash 1, Little Splash 2, Little Splash 3, Little Splash 4, Little Splash 5.

## Teen Adult 1

Source: [TeenAdult1.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/TeenAdult1.html)

SHA-256: `f72fb15eae1a7aca4db524bb09d28399ed16ed904812ebb8f71715ae936c7e46`

### 1. Enter and Exit Shallow Water

App ID: `TeenAdult1:1` · Short form: **Shallow entry & exit**

Back-page source: `TeenAdult1:1` — Enter and Exit Shallow Water

1. Foot first entry
2. Safe movement and control during entry and exit

### 2. Jump into Deep Water, Return and Exit

App ID: `TeenAdult1:2` · Short form: **Deep jump, return & exit**

Back-page source: `TeenAdult1:2` — Jump into Deep Water, Return and Exit

1. Foot-first entry
2. Controlled return to surface and safe return to side

### 3. Sideways Entry Wearing PFD

App ID: `TeenAdult1:3` · Short form: **PFD sideways entry**

Back-page source: `TeenAdult1:3` — Sideways Entry Wearing PFD

1. Appropriate PFD correctly donned and fastened on land
2. Controlled return to surface after entry

### 4. Tread water (30 Sec) Wearing PFD

App ID: `TeenAdult1:4` · Short form: **PFD tread 30s**

Back-page source: `TeenAdult1:4` — Tread Water (30 sec) Wearing PFD

1. Vertical body position
2. Mouth and nose above surface
3. Sculling action of hand generates support
4. Continuous supportive kick
5. Minimum time met

### 5. Hold Breath Underwater 5-10 sec.

App ID: `TeenAdult1:5` · Short form: **Breath hold 5–10s**

Back-page source: `TeenAdult1:5` — Hold Breath Underwater 5-10 sec

1. Entire body submerged
2. Minimum time met

### 6. Submerge and Exhale 5-10 Times

App ID: `TeenAdult1:6` · Short form: **Submerge/exhale ×5–10**

Back-page source: `TeenAdult1:6` — Submerge and Exhale 5-10 Times

1. Entire body submerged
2. Controlled exhalation underwater (5-10 times)

### 7. Open Eyes Underwater

App ID: `TeenAdult1:7` · Short form: **Open Eyes Underwater**

Back-page source: `TeenAdult1:7` — Open Eyes Underwater

1. Full face submerged with eyes open

### 8. Recover Object From Bottom in Chest-Deep Water

App ID: `TeenAdult1:8` · Short form: **Chest-deep object recovery**

Back-page source: `TeenAdult1:8` — Recover Object from Bottom in Chest-Deep Water

1. Face in water
2. Object recovered with hands and returned to surface

### 9. Wearing PFD, Jump into Deep Water, Tread 30 sec. and Swim/Kick on Back 5-10 m

App ID: `TeenAdult1:9` · Short form: **PFD: jump/back 5–10m**

Back-page source: `TeenAdult1:9` — Wearing PFD, Jump into Deep Water, Tread 30 sec and Swim/Kick on Back 5-10 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 10a. Float on Front

App ID: `TeenAdult1:10a` · Short form: **Front float**

Back-page source: `TeenAdult1:10` — Float on Front and Back

1. Float on front with face in water
2. Recovery from front and back floats

### 10b. Float on Back

App ID: `TeenAdult1:10b` · Short form: **Back float**

Back-page source: `TeenAdult1:10` — Float on Front and Back

1. Float on back with ears in water
2. Recovery from front and back floats

### 11. Roll Laterally Front to Back and Back to Front

App ID: `TeenAdult1:11` · Short form: **Front/back rolls**

Back-page source: `TeenAdult1:11` — Roll Laterally Front to Back and Back to Front

1. Begin in front or back float position
2. **Controlled lateral rollover:** roll front to back
3. **Controlled lateral rollover:** roll back to front
4. Body remains horizontal

### 12a. Glide on Front(3-5 m.)

App ID: `TeenAdult1:12-front-glide` · Short form: **Front glide 3–5m**

Back-page source: `TeenAdult1:12` — Glide on Front, Back and Side (3-5 m each)

1. **Front glide:** face in water
2. **Front glide:** underwater exhalation
3. **Front glide:** arms extended beyond head
4. Minimum distance completed

### 12b. Glide on Back (3-5 m.)

App ID: `TeenAdult1:12-back-glide` · Short form: **Back glide 3–5m**

Back-page source: `TeenAdult1:12` — Glide on Front, Back and Side (3-5 m each)

1. **Back Glide:** ears in water
2. **Back Glide:** arms by sides
3. Minimum distance completed

### 12c. Glide on Side (3-5 m.)

App ID: `TeenAdult1:12-side-glide` · Short form: **Side glide 3–5m**

Back-page source: `TeenAdult1:12` — Glide on Front, Back and Side (3-5 m each)

1. **Side glide:** body on side
2. **Side glide:** bottom arm extended beyond head
3. **Side glide:** top arm by side
4. **Side glide:** head resting on bottom arm
5. Minimum distance completed

### 13a. Flutter Kick on Front (10-15 m.)

App ID: `TeenAdult1:13a` · Short form: **Front flutter 10–15m**

Back-page source: `TeenAdult1:13` — Flutter Kick on Front, Back and Side (10-15 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 13b. Flutter Kick on Back (10-15 m.)

App ID: `TeenAdult1:13b` · Short form: **Back flutter 10–15m**

Back-page source: `TeenAdult1:13` — Flutter Kick on Front, Back and Side (10-15 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 13c. Flutter Kick on Side (10-15 m.)

App ID: `TeenAdult1:13c` · Short form: **Side flutter 10–15m**

Back-page source: `TeenAdult1:13` — Flutter Kick on Front, Back and Side (10-15 m each)

1. Appropriate streamlined body position
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
3. Minimum distance completed (in each position)

### 14. Whip kick in Vertical Position (15-30 sec.) with Aid

App ID: `TeenAdult1:14` · Short form: **Aided vertical whip 15–30s**

Back-page source: `TeenAdult1:14` — Whip Kick in Vertical Position (15-30 sec) with Aid

1. Kick in a vertical body position
2. Kick is simultaneous and symmetrical; heels recover towards body buttocks

### 15a. Front Crawl 10-15 m

App ID: `TeenAdult1:15-front-crawl` · Short form: **Front crawl 10–15m**

Back-page source: `TeenAdult1:15` — Front Crawl or Back Crawl 10-15 m

1. Minimum distance completed
2. **Front Crawl:** Body on front
3. **Front Crawl:** Alternate arm action
4. **Front Crawl:** Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
5. **Front Crawl:** Breathing with underwater exhalation

### 15b. Back Crawl 10-15 m

App ID: `TeenAdult1:15-back-crawl` · Short form: **Back crawl 10–15m**

Back-page source: `TeenAdult1:15` — Front Crawl or Back Crawl 10-15 m

1. Minimum distance completed
2. **Back Crawl:** Body on back
3. **Back Crawl:** Alternate arm action
4. **Back Crawl:** Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend
5. **Back Crawl:** Relaxed breathing

### 16. Interval Training: 4 x 9-12 m Flutter Kick with 10-15 sec Rests

App ID: `TeenAdult1:16` · Short form: **Flutter kick 4×9–12m**

Back-page source: `TeenAdult1:16` — Interval Training: 4 x 9-12 m Flutter Kick with 10-15 sec Rests

1. Repetitions completed
2. Propulsive, rhythmic flutter kick with alternate leg drive; slight knee bend

### 17. Water Smart Messages

App ID: `TeenAdult1:17` · Short form: **Water Smart Messages**

Back-page source: `TeenAdult1:17` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim with a Buddy:** Many drowning victims were swimming alone
3. **Wear a Lifejacket:** Most parents make kids wear a lifejacket but don't do it themselves
4. **Check the ice:** Learn how to measure ice thickness and then check it before you go
5. **Swim to Survive:** 2/3 of those who drown, do so within 15 m of safety
6. **Within Arms' Reach:** Toddlers and seniors are the most likely to drown in bathtubs

## Teen Adult 2

Source: [TeenAdult2.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/TeenAdult2.html)

SHA-256: `1999fa32e632d6713ae62891b76c73397b8bb6b2a935b2c895b25051a8abc18f`

### 1. Standing Dive into Deep Water

App ID: `TeenAdult2:1` · Short form: **Standing deep-water dive**

Back-page source: `TeenAdult2:1` — Standing Dive into Deep Water

1. Head-first entry from a standing position with controlled return to surface

### 2. Forward Roll Entry into Deep Water With and Without PFD

App ID: `TeenAdult2:2` · Short form: **Roll entry ±PFD**

Back-page source: `TeenAdult2:2` — Forward Roll Entry into Deep Water With and Without PFD

1. Controlled return to surface after entry

### 3. Tuck Jump (Cannonball) into Deep Water

App ID: `TeenAdult2:3` · Short form: **Deep-water tuck jump**

Back-page source: `TeenAdult2:3` — Tuck Jump (Cannonball) into Deep Water

1. Controlled return to surface after entry

### 4. Tread Water 1-2 min

App ID: `TeenAdult2:4` · Short form: **Tread Water 1-2 min**

Back-page source: `TeenAdult2:4` — Tread Water 1-2 min

1. Vertical body position
2. Mouth and nose above surface
3. Supportive andContinuous arm and leg action
4. Minimum time met

### 5. Handstand in Shallow Water

App ID: `TeenAdult2:5` · Short form: **Shallow handstand**

Back-page source: `TeenAdult2:5` — Handstand in Shallow Water

1. Hands on bottom; Feet off bottom
2. Head fully submerged

### 6. Front Somersault (in water)

App ID: `TeenAdult2:6` · Short form: **Front somersault**

Back-page source: `TeenAdult2:6` — Front Somersault (in water)

1. Forward roll with "head over heels" rotation

### 7. Swim Underwater 5-10 m

App ID: `TeenAdult2:7` · Short form: **Swim Underwater 5-10 m**

Back-page source: `TeenAdult2:7` — Swim Underwater 5-10 m

1. Body fully submerged throughout
2. Minimum distance completed

### 8. Canadian Swim to Survive Standard: Roll Entry into Deep Water, Tread 1 min, and Swim 50 m

App ID: `TeenAdult2:8` · Short form: **Swim to Survive 50m**

Back-page source: `TeenAdult2:8` — Canadian Swim to Survive Standard: Roll Entry into Deep Water, Tread 1 min, and Swim 50 m

1. Completion of skills in a continuous sequence
2. Distance and time requirements completed

### 9. Flutter Kick on Back 5 m; Reverse Direction and Flutter Kick on Front 5 m

App ID: `TeenAdult2:9` · Short form: **Back→front flutter 5m**

Back-page source: `TeenAdult2:9` — Flutter Kick on Back 5 m; Reverse Direction and Flutter Kick on Front 5 m

1. Appropriate streamlined body position
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend, feet relaxed
3. Controlled change from back to front
4. Minimum distance completed (in both positions)

### 10. Flutter Kick on Front 5 m; Reverse Direction and Flutter Kick on back 5 m

App ID: `TeenAdult2:10` · Short form: **Front→back flutter 5m**

Back-page source: `TeenAdult2:10` — Flutter Kick on Front 5 m; Reverse Direction and Flutter Kick on Back 5 m

1. Appropriate streamlined body position
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend, feet relaxed
3. Controlled change from front to back
4. Minimum distance completed (in both positions)

### 11. Whip Kick on back 10-15 m

App ID: `TeenAdult2:11` · Short form: **Whip Kick on back 10-15 m**

Back-page source: `TeenAdult2:11` — Whip Kick on Back 10-15 m

1. body on back
2. Kick is simultaneous and symmetrical; heels recover towards buttocks
3. Legs drive with knees apart; feet wider than knees
4. Minimum distance completed

### 12. Whip Kick on Front 10-15 m

App ID: `TeenAdult2:12` · Short form: **Front whip kick 10–15m**

Back-page source: `TeenAdult2:12` — Whip Kick on Front 10-15 m

1. body on front
2. Kick is simultaneous and symmetrical; heels recover towards buttocks
3. Legs drive with knees apart; feet wider than knees
4. Minimum distance completed

### 13. Breaststroke Arms Drill 10-15 m

App ID: `TeenAdult2:13` · Short form: **Breaststroke arms 10–15m**

Back-page source: `TeenAdult2:13` — Breastroke Arms Drill 10-15 m

1. Horizontal body position on front
2. Symmetrical arm pull towards chest; elbows higher than hands; symmetrical arm recovery beyond head
3. Regular breathing pattern with underwater exhalation
4. Minimum distance completed

### 14a. Front Crawl Crawl (25-50 m.)

App ID: `TeenAdult2:14-front-crawl` · Short form: **Front crawl 25–50m**

Back-page source: `TeenAdult2:14` — Front Crawl and Back Crawl (25-50 m each)

1. Minimum distance completed
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
5. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 14b. Back Crawl (25-50 m.)

App ID: `TeenAdult2:14-back-crawl` · Short form: **Back crawl 25–50m**

Back-page source: `TeenAdult2:14` — Front Crawl and Back Crawl (25-50 m each)

1. Minimum distance completed
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
5. **Back Crawl:** Relaxed breathing

### 15. Interval Training: 4 x 25 m Flutter Kick with 15-20 sec Rests

App ID: `TeenAdult2:15` · Short form: **Flutter kick 4×25m**

Back-page source: `TeenAdult2:15` — Interval Training: 4 x 25 m Flutter Kick with 15-20 sec Rests

1. Repetitions completed
2. Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed

### 16. Interval Training: 4 x 25 m Front or Back Crawl with 15-20 sec Rests

App ID: `TeenAdult2:16` · Short form: **Front/back crawl 4×25m**

Back-page source: `TeenAdult2:16` — Interval Training: 4 x 25 m Front or Back Crawl 10-15 sec Rests

1. Repetitions completed
2. Front Crawl/Back Crawl must sees (see item 14)

### 17. Sprint Front Crawl 25 m

App ID: `TeenAdult2:17` · Short form: **Sprint Front Crawl 25 m**

Back-page source: `TeenAdult2:17` — Sprint Front Crawl 25 m

1. Distance completed as fast as possible

### 18a. Water Smart Messages - Swim with a Buddy

App ID: `TeenAdult2:18a` · Short form: **Swim with a Buddy**

Back-page source: `TeenAdult2:18` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim with a Buddy:** Many drowning victims were swimming alone

### 18b. Water Smart Messages - Wear a Lifejacket

App ID: `TeenAdult2:18b` · Short form: **Wear a Lifejacket**

Back-page source: `TeenAdult2:18` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Wear a Lifejacket:** Most victims in boat fatalities were not wearing a lifejacket

### 18c. Water Smart Messages - Check the ice

App ID: `TeenAdult2:18c` · Short form: **Check the Ice**

Back-page source: `TeenAdult2:18` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Check the ice:** Dog walkers need to be careful
3. **Check the ice:** Many drown trying to save the dog (the dog usually survives)

### 18d. Water Smart Messages - Swim to Survive

App ID: `TeenAdult2:18d` · Short form: **Swim to Survive**

Back-page source: `TeenAdult2:18` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim to Survive:** The majority of drowning victims did not even intend to get into the water

### Back-page requirements with no matching front column

- `TeenAdult2:18:7` **Within Arms' Reach:** Use a back yard pool gate that self-closes and latches

## Teen Adult 3

Source: [TeenAdult3.html](../../COB-Aquatics-SPA/frontend/src/shared/attendance-print/templates/TeenAdult3.html)

SHA-256: `09560423fe559062a9c0007e0face9beb2b037f6e17e2cc338d163e1ff6deba1`

### 1. Shallow Dive into Deep Water

App ID: `TeenAdult3:1` · Short form: **Shallow dive**

Back-page source: `TeenAdult3:1` — Shallow Dive into Deep Water

1. Head-first entry following a path close to the surface: surface quickly

### 2. Stride Entry into Deep Water

App ID: `TeenAdult3:2` · Short form: **Deep-water stride entry**

Back-page source: `TeenAdult3:2` — Stride Entry into Deep Water

1. Foot-first Entry with Head above surface

### 3. Compact Jump into Deep Water

App ID: `TeenAdult3:3` · Short form: **Deep-water compact jump**

Back-page source: `TeenAdult3:3` — Compact Jump into Deep Water

1. Vertical, Feet-first entry
2. Crossed legs, arms tight to body
3. Airway protected

### 4. Legs-only Surface support 30-60 sec

App ID: `TeenAdult3:4` · Short form: **Legs-only support 30–60s**

Back-page source: `TeenAdult3:4` — Legs-only Surface Support (30-60 sec

1. Vertical body position
2. Mouth and nose above surface
3. Recognizable supportive lifesaving kick
4. Minimum time met

### 5. Back Somersault (in water)

App ID: `TeenAdult3:5` · Short form: **Back Somersault (in water)**

Back-page source: `TeenAdult3:5` — Back Somersault (in water)

1. Backward roll with "Head over heels" rotation

### 6. Swim Underwater 5-10 m to Recover Object

App ID: `TeenAdult3:6` · Short form: **Underwater pickup 5–10m**

Back-page source: `TeenAdult3:6` — Swim Underwater 5-10 m to Recover Object

1. Body fully submerged throughout
2. Minimum distance completed
3. Object recovered and returned to surface

### 7. Eggbeater Kick on Back or Scissor Kick on Side 10-15 m

App ID: `TeenAdult3:7` · Short form: **Eggbeater/scissor 10–15m**

Back-page source: `TeenAdult3:7` — Eggbeater Kick on Back or Scissor Kick on Side 10-15 m

1. Minimum distance completed
2. **Eggbeater Kick:** Body on back
3. **Eggbeater Kick:** Semi-circular leg drive toward mid-line of body; alternate drive and recovery
4. **Eggbeater Kick:** Drive with knees apart; feet wider than knees
5. **Scissor Kick:** Body on side
6. **Scissor Kick:** Both legs kick at same time and recover at the same time
7. **Scissor Kick:** one leg extends forward and other leg extends back to begin drive
8. **Scissor Kick:** Legs squeeze together in scissor-like action for drive phase

### 8. Breaststroke 25-50 m

App ID: `TeenAdult3:8` · Short form: **Breaststroke 25-50 m**

Back-page source: `TeenAdult3:8` — Breaststroke 25-50 m

1. Horizontal body position on front
2. Symmetrical arm pull towards chest; elbows higher than hands; symmetrical arm recovery beyond head
3. Regular breathing pattern with underwater exhalation
4. Kick is simultaneous and symmetrical; heels recover towards buttocks
5. Legs drive with knees apart; feet wider than knees
6. Pull-breathe-kick-stretch sequence into full streamlined extension
7. Minimum distance completed

### 9a. Front Crawl (50-100 m.)

App ID: `TeenAdult3:9-front-crawl` · Short form: **Front crawl 50–100m**

Back-page source: `TeenAdult3:9` — Front Crawl and Back Crawl (50-100 m. each)

1. Minimum distance completed
2. **Front Crawl:** Body on front; streamlined and horizontal
3. **Front Crawl:** Alternate arm action with overarm recovery and underwater pull below body
4. **Front Crawl:** Hand enters beyond shoulder, pull past hip
5. **Front Crawl:** Elbow higher than hand during recovery
6. **Front Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Front Crawl:** Regular breathing pattern to side with underwater exhalation

### 9b. Back Crawl (50-100 m.)

App ID: `TeenAdult3:9-back-crawl` · Short form: **Back crawl 50–100m**

Back-page source: `TeenAdult3:9` — Front Crawl and Back Crawl (50-100 m. each)

1. Minimum distance completed
2. **Back Crawl:** Body on back; streamlined position
3. **Back Crawl:** Alternate arm action with overarm recovery and underwater pull
4. **Back Crawl:** Hand drive finishes at thigh; recovery above surface with elbows straight; entry beyond shoulder
5. **Back Crawl:** Drive with bent arm pull
6. **Back Crawl:** Continuous, rhythmic flutter kick initiated from hips; slight knee bend; feet relaxed
7. **Back Crawl:** Relaxed breathing

### 10. Head-up Front Crawl 10-15 m

App ID: `TeenAdult3:10` · Short form: **Head-up crawl 10–15m**

Back-page source: `TeenAdult3:10` — Head up Front Crawl 10-15 m

1. Eyes above surface
2. Minimum distance completed

### 11. Interval Training: 4 x 50 m Front or Back Crawl or Breaststroke with 30 sec Rests

App ID: `TeenAdult3:11` · Short form: **Crawl/breaststroke 4×50m**

Back-page source: `TeenAdult3:11` — Interval Training: 4 x 50 m Front or Back Crawl or Breaststroke with 30 sec Rests

1. Repetitions completed
2. Front Crawl/Back Crawl/Breaststroke must sees (see items 8 and 9)

### 12. Sprint (25-50 m) Front Crawl, Back Crawl or Breaststroke

App ID: `TeenAdult3:12` · Short form: **F/B/br. sprint 25–50m**

Back-page source: `TeenAdult3:12` — Sprint Front Crawl, Back Crawl or Breaststroke 25-50 m

1. Distance completed as fast as possible

### 13. Workout 300 m: 50 m Warm-up (choice of Strokes); 4 x 25 m Front Crawl with 15 sec Rests; 4 x25 m Back Crawl with 15 sec Rests; 50 m Cool-down (Choice of Strokes)

App ID: `TeenAdult3:13` · Short form: **Workout 300m**

Back-page source: `TeenAdult3:13` — Workout 300 m: 50 m Warm-up (choice of strokes); 4 x 25 Front Crawl with 15 sec Rests; 4 x 25 m Back Crawl with 15 sec Rests; 50 m Cool-down (choice of strokes)

1. Workout completed

### 14a. Water Smart Messages - Swim with a Buddy

App ID: `TeenAdult3:14a` · Short form: **Swim with a Buddy**

Back-page source: `TeenAdult3:14` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Swim with a Buddy:** Many drowning victims were swimming alone

### 14b. Water Smart Messages Wear a Lifejacket

App ID: `TeenAdult3:14b` · Short form: **Wear a Lifejacket**

Back-page source: `TeenAdult3:14` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Wear a Lifejacket:** Lifejacket is like a seat belt
3. **Wear a Lifejacket:** Wear it before a crash

### 14c. Water Smart Messages - Check the ice

App ID: `TeenAdult3:14c` · Short form: **Check the Ice**

Back-page source: `TeenAdult3:14` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Check the ice:** Recreational snowmobiler's are at high-risk for drowning
3. **Check the ice:** Check the ice

### 14d. Water Smart Messages - Within Arms

App ID: `TeenAdult3:14d` · Short form: **Within Arms' Reach**

Back-page source: `TeenAdult3:14` — Water Smart Messages

1. Participation in a water activity reinforcing the water smart messages:
2. **Within Arms' Reach:** Drain standing water

### Back-page requirements with no matching front column

- `TeenAdult3:14:7` **Swim to Survive:** The majority of drowning victims did not even intend to get into the water
