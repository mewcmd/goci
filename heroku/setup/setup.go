//setup provides functionality to build the Go distribution from source
package setup

import (
	"errors"
	"fmt"
	"github.com/zeebo/goci/environ"
	"github.com/zeebo/goci/tarball"
	"net/http"
	"os"
	"runtime"
	"strings"
	fp "path/filepath"
)

type LocalWorld interface {
	Environ() []string
	Exists(string) bool
	LookPath(string) (string, error)
	MkdirAll(string, os.FileMode) error
	Make(environ.Command) environ.Proc
}

var World LocalWorld = environ.New()

func InstallGo(dir string) (bin string, err error) {
	dir = fp.Clean(dir)

	vers := runtime.Version()
	src := fmt.Sprintf("http://go.googlecode.com/files/%s.src.tar.gz", vers)

	resp, err := http.Get(src)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//make the goroot
	if err = World.MkdirAll(dir, 0666); err != nil {
		return
	}

	err = tarball.Extract(resp.Body, dir)
	if err != nil {
		return
	}

	bin = fp.Join(dir, "go", "bin")
	return
}

func InstallVCS(dir string) (bin string, err error) {
	//check for our dist directory
	distdir := fp.Join("heroku", "dist")
	if !World.Exists(distdir) {
		err = errors.New("unable to find dist directory: " + distdir)
		return
	}

	cmds := []string{"python2.7", "python", "basename"}
	for _, cmd := range cmds {
		if _, err = World.LookPath(cmd); err != nil {
			return
		}
	}

	dir = fp.Join(dir, "venv")

	vcs_inst_cmd := fmt.Sprintf(`
		python "%s" --python python2.7 --distribute --never-download %s
		. %s
		pip install --use-mirrors mercurial
		pip install --use-mirrors bzr
	`,
		fp.Join(distdir, "virtualenv-1.7", "virtualenv.py"),
		fp.Dir(dir),
		fp.Join(dir, "activate"),
	)

	path, err := World.LookPath("bash")
	if err != nil {
		return
	}
	p := World.Make(environ.Command{
		R:    strings.NewReader(vcs_inst_cmd),
		Path: path,
		Args: []string{path},
		Env:  World.Environ(),
	})
	err, _ = p.Run()
	if err != nil {
		return
	}
	bin = fp.Join(dir, "bin")
	return
}
