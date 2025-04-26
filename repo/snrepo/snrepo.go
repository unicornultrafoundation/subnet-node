package snrepo

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	measure "github.com/ipfs/go-ds-measure"
	lockfile "github.com/ipfs/go-fs-lock"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

const LockFile = "repo.lock"

const (
	swarmKeyFile = "swarm.key"
)

var (
	packageLock sync.Mutex

	onlyOne repo.OnlyOne
)

type SNRepo struct {
	// has Close been called already
	closed bool
	// path is the file-system path
	path           string
	configFilePath string
	lockfile       io.Closer
	config         *config.C
	ds             repo.Datastore
}

var _ repo.Repo = (*SNRepo)(nil)

func Open(repoPath string, userConfigFilePath *string) (repo.Repo, error) {
	fn := func() (repo.Repo, error) {
		return open(repoPath, userConfigFilePath)
	}
	return onlyOne.Open(repoPath, fn)
}

func open(repoPath string, userConfigFilePath *string) (repo.Repo, error) {
	packageLock.Lock()
	defer packageLock.Unlock()
	r, err := newSNRepo(repoPath, userConfigFilePath)
	if err != nil {
		return nil, err
	}

	r.lockfile, err = lockfile.Lock(r.path, LockFile)
	if err != nil {
		return nil, err
	}
	keepLocked := false
	defer func() {
		// unlock on error, leave it locked on success
		if !keepLocked {
			r.lockfile.Close()
		}
	}()

	if err := r.openDatastore(); err != nil {
		return nil, err
	}

	return r, nil
}

func newSNRepo(rpath string, userConfigFilePath *string) (*SNRepo, error) {
	expPath, err := fsutil.ExpandHome(filepath.Clean(rpath))
	if err != nil {
		return nil, err
	}

	l := logrus.New()
	l.Out = os.Stdout

	configPath := filepath.Join(expPath, "config.yaml")
	if *userConfigFilePath != "" {
		configPath = *userConfigFilePath
	}
	c := config.NewC(l)
	err = c.Load(configPath)
	if err != nil {
		log.Printf("failed to load config: %s", err)
		os.Exit(1)
	}

	return &SNRepo{
		configFilePath: configPath,
		path:           expPath,
		config:         c,
	}, nil
}

func (r *SNRepo) Config() *config.C {
	return r.config
}

func (r *SNRepo) Path() string {
	return r.path
}

func (r *SNRepo) Datastore() repo.Datastore {
	packageLock.Lock()
	d := r.ds
	packageLock.Unlock()
	return d
}

func (r *SNRepo) SwarmKey() ([]byte, error) {
	repoPath := filepath.Clean(r.path)
	spath := filepath.Join(repoPath, swarmKeyFile)

	f, err := os.Open(spath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func (r *SNRepo) Close() error {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return errors.New("repo is closed")
	}

	if err := r.ds.Close(); err != nil {
		return err
	}

	r.closed = true
	return r.lockfile.Close()
}

// openDatastore returns an error if the config file is not present.
func (r *SNRepo) openDatastore() error {
	defaultCfg := map[string]any{
		"type": "mount",
		"mounts": []any{
			map[string]any{
				"mountpoint": "/",
				"type":       "measure",
				"prefix":     "leveldb.datastore",
				"child": map[string]any{
					"type":        "levelds",
					"path":        "datastore",
					"compression": "none",
				},
			},
		},
	}
	dsc, err := AnyDatastoreConfig(r.config.GetMap("datastore.spec", defaultCfg))
	if err != nil {
		return err
	}

	d, err := dsc.Create(r.path)
	if err != nil {
		return err
	}
	r.ds = d

	// Wrap it with metrics gathering
	prefix := "ipsn.nsrepo.datastore"
	r.ds = measure.New(prefix, r.ds)
	return nil
}

// IsInitialized returns true if the repo is initialized at provided |path|.
func IsInitialized(path string) bool {
	// packageLock is held to ensure that another caller doesn't attempt to
	// Init or Remove the repo while this call is in progress.
	packageLock.Lock()
	defer packageLock.Unlock()

	return configIsInitialized(path)
}

// configIsInitialized returns true if the repo is initialized at
// provided |path|.
func configIsInitialized(path string) bool {
	expPath, err := fsutil.ExpandHome(filepath.Clean(path))
	if err != nil {
		return false
	}

	l := logrus.New()
	l.Out = os.Stdout

	configPath := filepath.Join(expPath, "config.yaml")
	return fsutil.FileExists(configPath)
}
