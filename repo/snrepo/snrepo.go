package snrepo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ipfs/boxo/keystore"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/misc/fsutil"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

const (
	swarmKeyFile = "swarm.key"
)

var (

	// packageLock must be held to while performing any operation that modifies an
	// FSRepo's state field. This includes Init, Open, Close, and Remove.
	packageLock sync.Mutex

	// onlyOne keeps track of open FSRepo instances.
	//
	// TODO: once command Context / Repo integration is cleaned up,
	// this can be removed. Right now, this makes ConfigCmd.Run
	// function try to open the repo twice:
	//
	//     $ ipfs daemon &
	//     $ ipfs config foo
	//
	// The reason for the above is that in standalone mode without the
	// daemon, `ipfs config` tries to save work by not building the
	// full IpfsNode, but accessing the Repo directly.
	onlyOne repo.OnlyOne
)

type FNRepo struct {
	// has Close been called already
	closed bool
	// path is the file-system path
	path string
	// Path to the configuration file that may or may not be inside the FSRepo
	// path (see config.Filename for more details).
	configFilePath string
	// lockfile is the file system lock to prevent others from opening
	// the same fsrepo path concurrently
	lockfile              io.Closer
	config                *config.C
	userResourceOverrides rcmgr.PartialLimitConfig
	ds                    repo.Datastore
	keystore              keystore.Keystore
}

var _ repo.Repo = (*FNRepo)(nil)

// Open the FSRepo at path. Returns an error if the repo is not
// initialized.
func Open(repoPath string) (repo.Repo, error) {
	fn := func() (repo.Repo, error) {
		return open(repoPath, "")
	}
	return onlyOne.Open(repoPath, fn)
}

// OpenWithUserConfig is the equivalent to the Open function above but with the
// option to set the configuration file path instead of using the default.
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

func newFSRepo(rpath string, userConfigFilePath string) (*FNRepo, error) {
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

	return &FNRepo{
		configFilePath: userConfigFilePath,
		path:           expPath,
		config:         c,
	}, nil
}

// Close closes the FSRepo, releasing held resources.
func (r *FNRepo) Close() error {
	return nil
}

func (r *FNRepo) Config() *config.C {
	return r.config
}

// Datastore returns a repo-owned datastore. If FSRepo is Closed, return value
// is undefined.
func (r *FNRepo) Datastore() repo.Datastore {
	packageLock.Lock()
	d := r.ds
	packageLock.Unlock()
	return d
}

func (r *FNRepo) SwarmKey() ([]byte, error) {
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
