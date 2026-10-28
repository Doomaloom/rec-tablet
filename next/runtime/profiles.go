package runtime

import (
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
	Key        string `json:"key"`
	DatasetID  string `json:"datasetId"`
	Instructor string `json:"instructor"`
	Path       string `json:"path"`
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

func (m *profileManager) packagePath(pkg transfer.ClassPackage) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := profileKey(pkg.DatasetID, pkg.Instructor)
	if path, ok := m.pathForLocked(key); ok {
		return path
	}
	digest := sha256.Sum256([]byte(key))
	return filepath.Join(filepath.Dir(m.base), filepath.Base(m.base)+".sessions", hex.EncodeToString(digest[:])+".sqlite3")
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
	if _, exists := m.pathForLocked(profileKey(pkg.DatasetID, pkg.Instructor)); exists {
		if _, err := os.Stat(path); err != nil {
			return savedProfile{}, errors.New("saved instructor database is missing; refusing to create a replacement")
		}
	}
	target, err := store.Open(path)
	if err != nil {
		return savedProfile{}, err
	}
	if _, err = target.Import(pkg); err != nil {
		target.Close()
		return savedProfile{}, err
	}
	if err = target.Close(); err != nil {
		return savedProfile{}, err
	}
	profile, err := m.registerLocked(pkg.DatasetID, pkg.Instructor, path)
	return profile, err
}

func (m *profileManager) packagePathLocked(pkg transfer.ClassPackage) string {
	key := profileKey(pkg.DatasetID, pkg.Instructor)
	if path, ok := m.pathForLocked(key); ok {
		return path
	}
	digest := sha256.Sum256([]byte(key))
	return filepath.Join(filepath.Dir(m.base), filepath.Base(m.base)+".sessions", hex.EncodeToString(digest[:])+".sqlite3")
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
