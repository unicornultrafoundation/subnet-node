package snrepo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/misc/fsutil"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

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

func Open(repoPath string) (repo.Repo, error) {
	fn := func() (repo.Repo, error) {
		return open(repoPath, "")
	}
	return onlyOne.Open(repoPath, fn)
}

func OpenWithUserConfig(repoPath string, userConfigFilePath string) (repo.Repo, error) {
	fn := func() (repo.Repo, error) {
		return open(repoPath, userConfigFilePath)
	}
	return onlyOne.Open(repoPath, fn)
}

func open(repoPath string, userConfigFilePath string) (repo.Repo, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	r, err := newFSRepo(repoPath, userConfigFilePath)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func newFSRepo(rpath string, userConfigFilePath string) (*SNRepo, error) {
	expPath, err := fsutil.ExpandHome(filepath.Clean(rpath))
	if err != nil {
		return nil, err
	}

	l := logrus.New()
	l.Out = os.Stdout

	c := config.NewC(l)
	err = c.Load(userConfigFilePath)
	if err != nil {
		fmt.Printf("failed to load config: %s", err)
		os.Exit(1)
	}

	return &SNRepo{
		configFilePath: userConfigFilePath,
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
