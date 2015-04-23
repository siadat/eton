package main

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"log"
	"os"
	"os/signal"
	//"path/filepath"
)

type HelloFs struct {
	pathfs.FileSystem
}

func (me *HelloFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	log.Printf("GetAttr for %v\n", name)
	switch name {
	case "":
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	default:
		attr := findAttributeByAliasOrID(globalDB, name)
		size := 0
		if attr.GetID() > 0 {
			size = len(attr.GetValue())
		}
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644,
			Size: uint64(size),
		}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (me *HelloFs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	log.Printf("OpenDir for %v\n", name)
	if name == "" {
		//c = []fuse.DirEntry{{Name: "file.txt", Mode: fuse.S_IFREG}}
		attrs := listWithFilters(globalDB, globalOpts)
		c = make([]fuse.DirEntry, len(attrs), len(attrs))

		for i, attr := range attrs {
			var d fuse.DirEntry
			d.Name = attr.GetIdentifier()
			//log.Println(d.Name)
			d.Mode = fuse.S_IFREG | 0644
			c[i] = fuse.DirEntry(d)
			//c = append(c, )
		}

		return c, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (me *HelloFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	log.Printf("Open for %v\n", name)
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	attr := findAttributeByAliasOrID(globalDB, name)
	if attr.GetID() <= 0 {
		return nil, fuse.ENOENT
	}
	bytes := []byte(attr.GetValue())
	return nodefs.NewDataFile(bytes), fuse.OK
}

func Mount(mountpoint string) {

	log.Println("NOTE: This is just an experiment")
	log.Println("mounting on", mountpoint)
	if _, err := os.Stat(mountpoint); err == nil {
		log.Println("directory exists:", mountpoint)
		log.Println("move directory and try again")
		return
	} else {
		os.Mkdir(mountpoint, os.ModeDir)
	}

	nfs := pathfs.NewPathNodeFs(&HelloFs{FileSystem: pathfs.NewDefaultFileSystem()}, nil)
	server, _, err := nodefs.MountRoot(mountpoint, nfs.Root(), nil)
	check(err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	defer func() {
		server.Unmount()
		//mountpointAbsolutepath, _ := filepath.Abs(mountpoint)
		os.Remove(mountpoint)
		log.Println("Removed mount point", mountpoint)
	}()

	go func() {
		for _ = range c {
			// CTRL-c
			// Do nothing, this will continue executing the rest of the code
			server.Unmount()
		}
	}()

	server.Serve()
}
