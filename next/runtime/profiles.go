package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
)

type savedProfile struct {
	InstructorID string `json:"instructorId,omitempty"`
	Key          string `json:"key"`
	DatasetID    string `json:"datasetId"`
	Instructor   string `json:"instructor"`
	Path         string `json:"path"`
}

type profileIndex struct {
	Active      string         `json:"active"`
	Profiles    []savedProfile `json:"profiles"`
	ShareServer string         `json:"shareServer,omitempty"`
}

func (m *profileManager) shareServer(fallback string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if strings.TrimSpace(m.index.ShareServer) != "" {
		return m.index.ShareServer
	}
	return fallback
}

func (m *profileManager) setShareServer(value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	old := m.index.ShareServer
	m.index.ShareServer = strings.TrimRight(strings.TrimSpace(value), "/")
	if err := m.save(); err != nil {
		m.index.ShareServer = old
		return err
	}
	return nil
}

type profileManager struct {
	mu    sync.Mutex
	base  string
	index profileIndex
	// recovered is set when a listed instructor's database file was gone and a
	// new one was made, so the next login can say so.
	recovered bool
}

func newProfileManager(base string) (*profileManager, error) {
	manager := &profileManager{base: base}
	data, err := os.ReadFile(base + ".profiles.json")
	if err == nil {
		if err := json.Unmarshal(data, &manager.index); err != nil {
			return nil, errors.New("saved session index is invalid; preserve the file and contact support")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return manager, nil
}

func profileKey(dataset, instructor string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(dataset + "\x00" + instructor))
}

func (m *profileManager) register(dataset, instructor, path string) (savedProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.registerLocked(dataset, instructor, path)
}

func (m *profileManager) registerLocked(dataset, instructor, path string) (savedProfile, error) {
	if strings.TrimSpace(dataset) == "" || strings.TrimSpace(instructor) == "" {
		return savedProfile{}, nil
	}
	key := profileKey(dataset, instructor)
	for i := range m.index.Profiles {
		if m.index.Profiles[i].Key == key {
			if path == m.base && m.index.Profiles[i].Path != m.base {
				return m.index.Profiles[i], nil
			}
			m.index.Profiles[i].Path = path
			return m.index.Profiles[i], m.save()
		}
	}
	profile := savedProfile{Key: key, DatasetID: dataset, Instructor: instructor, Path: path}
	m.index.Profiles = append(m.index.Profiles, profile)
	return profile, m.save()
}

func (m *profileManager) registerStore(s *store.Store) error {
	dataset, err := s.Meta("dataset_id")
	if err != nil {
		return err
	}
	instructor, err := s.Meta("instructor")
	if err != nil {
		return err
	}
	id, err := s.SyncInstructorID()
	if err != nil {
		return err
	}
	if id != "" {
		m.mu.Lock()
		defer m.mu.Unlock()
		key := profileKey(dataset, "id:"+id)
		for i := range m.index.Profiles {
			p := &m.index.Profiles[i]
			if p.Path == s.DatabasePath() {
				p.Key = key
				p.InstructorID = id
				p.Instructor = instructor
				return m.save()
			}
		}
		m.index.Profiles = append(m.index.Profiles, savedProfile{Key: key, DatasetID: dataset, Instructor: instructor, InstructorID: id, Path: s.DatabasePath()})
		return m.save()
	}
	_, err = m.register(dataset, instructor, s.DatabasePath())
	return err
}

func (m *profileManager) pathFor(key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pathForLocked(key)
}

func (m *profileManager) pathForLocked(key string) (string, bool) {
	for _, profile := range m.index.Profiles {
		if profile.Key == key {
			return profile.Path, true
		}
	}
	return "", false
}

func packageKey(pkg transfer.ClassPackage) string {
	if pkg.Credentials != nil && pkg.Sync != nil {
		return profileKey(pkg.DatasetID, "id:"+pkg.Sync.InstructorID)
	}
	return profileKey(pkg.DatasetID, pkg.Instructor)
}
func (m *profileManager) packagePath(pkg transfer.ClassPackage) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.packagePathLocked(pkg)
}

func (m *profileManager) activePath() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if path, ok := m.pathForLocked(m.index.Active); ok {
		return path
	}
	return m.base
}

func (m *profileManager) importPackage(pkg transfer.ClassPackage) (savedProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := m.packagePathLocked(pkg)
	if m.existingPackageLocked(pkg) {
		// The saved file is gone, so there is nothing left to protect: start a
		// new one from the class package instead of blocking this instructor.
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			m.recovered = true
		} else if err != nil {
			return savedProfile{}, err
		}
	}
	target, err := store.Open(path)
	if err != nil {
		return savedProfile{}, err
	}
	if err = target.ValidateSyncProfile(pkg); err != nil {
		target.Close()
		if legacy, ok := m.legacyPathLocked(pkg); !ok || legacy != path {
			return savedProfile{}, err
		}
		// The name-keyed profile holds another class scope. Leave it as it is and
		// give this canonical instructor a separate database.
		path = m.freshPathLocked(packageKey(pkg))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return savedProfile{}, err
		}
		if target, err = store.Open(path); err != nil {
			return savedProfile{}, err
		}
		if err = target.ValidateSyncProfile(pkg); err != nil {
			target.Close()
			return savedProfile{}, err
		}
	}
	if _, err = target.Import(pkg); err != nil {
		target.Close()
		return savedProfile{}, err
	}
	if err = target.ConfigureSync(pkg); err != nil {
		target.Close()
		return savedProfile{}, err
	}
	if err = target.Close(); err != nil {
		return savedProfile{}, err
	}
	if pkg.Credentials == nil || pkg.Sync == nil {
		return m.registerLocked(pkg.DatasetID, pkg.Instructor, path)
	}
	key := packageKey(pkg)
	for i := range m.index.Profiles {
		p := &m.index.Profiles[i]
		if p.Key == key || (p.Path == path && p.DatasetID == pkg.DatasetID) {
			p.Key = key
			p.Instructor = pkg.Instructor
			p.InstructorID = pkg.Sync.InstructorID
			return *p, m.save()
		}
	}
	profile := savedProfile{Key: key, DatasetID: pkg.DatasetID, Instructor: pkg.Instructor, InstructorID: pkg.Sync.InstructorID, Path: path}
	m.index.Profiles = append(m.index.Profiles, profile)
	return profile, m.save()
}

func (m *profileManager) packagePathLocked(pkg transfer.ClassPackage) string {
	key := packageKey(pkg)
	// Prefer canonical identity; migrate the old name profile in place once.
	if path, ok := m.pathForLocked(key); ok {
		return path
	}
	if path, ok := m.legacyPathLocked(pkg); ok {
		return path
	}
	return m.freshPathLocked(key)
}

// legacyPathLocked finds the name-keyed profile a bound package may migrate.
func (m *profileManager) legacyPathLocked(pkg transfer.ClassPackage) (string, bool) {
	if _, ok := m.pathForLocked(packageKey(pkg)); ok || pkg.Credentials == nil {
		return "", false
	}
	for _, p := range m.index.Profiles {
		if p.InstructorID == "" && p.Key == profileKey(pkg.DatasetID, pkg.Instructor) {
			return p.Path, true
		}
	}
	return "", false
}

func (m *profileManager) freshPathLocked(key string) string {
	digest := sha256.Sum256([]byte(key))
	path := filepath.Join(filepath.Dir(m.base), filepath.Base(m.base)+".sessions", hex.EncodeToString(digest[:])+".sqlite3")
	// An in-place canonical migration may retain the former name-hashed path.
	// A later unbound package must never reopen that bound database by accident.
	for _, p := range m.index.Profiles {
		if p.Path == path && p.Key != key {
			digest = sha256.Sum256([]byte("unbound:" + key))
			path = filepath.Join(filepath.Dir(m.base), filepath.Base(m.base)+".sessions", hex.EncodeToString(digest[:])+".sqlite3")
			break
		}
	}
	return path
}

func (m *profileManager) registerPackage(pkg transfer.ClassPackage, path string) (savedProfile, error) {
	return m.register(pkg.DatasetID, pkg.Instructor, path)
}

func (m *profileManager) activate(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.pathForLocked(key); !ok {
		return errors.New("saved session is unavailable")
	}
	old := m.index.Active
	m.index.Active = key
	if err := m.save(); err != nil {
		m.index.Active = old
		return err
	}
	return nil
}

func (m *profileManager) rows() []struct{ Key, Label string } {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := make([]struct{ Key, Label string }, 0, len(m.index.Profiles))
	for _, profile := range m.index.Profiles {
		rows = append(rows, struct{ Key, Label string }{profile.Key, profile.Instructor + " · " + profile.DatasetID})
	}
	return rows
}

func (m *profileManager) save() error {
	data, err := json.MarshalIndent(m.index, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.base + ".profiles.json.tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, m.base+".profiles.json")
}

func (m *profileManager) lastProfile() (savedProfile, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, profile := range m.index.Profiles {
		if profile.Key == m.index.Active {
			return profile, true
		}
	}
	return savedProfile{}, false
}

// takeRecovered reports, once, that a missing instructor database was replaced.
func (m *profileManager) takeRecovered() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	was := m.recovered
	m.recovered = false
	return was
}

func (m *profileManager) existingPackageLocked(pkg transfer.ClassPackage) bool {
	key := packageKey(pkg)
	for _, p := range m.index.Profiles {
		if p.Key == key || (pkg.Credentials != nil && p.InstructorID == "" && p.Key == profileKey(pkg.DatasetID, pkg.Instructor)) {
			return true
		}
	}
	return false
}
func (m *profileManager) retryUploads(ctx context.Context, active string) {
	_ = m.retryUploadsSummary(ctx, active)
}
