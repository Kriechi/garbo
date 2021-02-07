package main

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/nwaples/rardecode"

	"github.com/mholt/archiver/v3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// UIFile is a helper struct to abstract the archive implementation and only extract the info for the UI
type UIFile struct {
	Name  string
	Size  uint64
	IsDir bool
}

// FileItem represents a single file or directory in the archive tree UI
type FileItem struct {
	NameLabel     *widget.Label
	SizeLabel     *widget.Label
	ExtractButton *widget.Button
}

func walkArchive(filePath string, treeData map[string]UIFile, dirData map[string][]UIFile) error {
	err := archiver.Walk(filePath, func(f archiver.File) error {
		var uif UIFile
		switch h := f.Header.(type) { // ref: https://github.com/mholt/archiver/issues/214
		case zip.FileHeader:
			uif = UIFile{Name: h.Name, Size: uint64(f.Size()), IsDir: f.IsDir()}
		case *tar.Header:
			uif = UIFile{Name: h.Name, Size: uint64(f.Size()), IsDir: f.IsDir()}
		case *rardecode.FileHeader:
			uif = UIFile{Name: h.Name, Size: uint64(f.Size()), IsDir: f.IsDir()}
		default:
			uif = UIFile{Name: f.Name(), Size: uint64(f.Size()), IsDir: f.IsDir()}
		}

		if strings.HasPrefix(path.Base(uif.Name), "._") {
			// hide weird macOS hidden files used for extended attributes
			return nil
		}

		dir, _ := path.Split(strings.TrimRight(uif.Name, string(os.PathSeparator)))
		if _, ok := dirData[dir]; !ok {
			dirData[dir] = []UIFile{}
		}
		dirData[dir] = append(dirData[dir], uif)
		treeData[uif.Name] = uif
		return nil
	})
	return err
}

func extractFile(filePath string, fileToExtract UIFile) {
	// "file" in the sense that one could also extract a directory and all its contents

	dialogSize := fyne.NewSize(700, 500)

	dialogWindow := fyne.CurrentApp().NewWindow("Select destination...")
	dialogWindow.Resize(dialogSize)
	dialogWindow.CenterOnScreen()
	dialogWindow.Show()

	dialogCallback := func(list fyne.ListableURI, err error) {
		dialogWindow.Close()
		if list == nil {
			return
		}
		if err != nil {
			log.Fatal(list, err)
		}

		destination := strings.TrimPrefix(list.String(), "file://")
		err = archiver.Extract(filePath, fileToExtract.Name, destination)
		if err != nil {
			log.Fatal(err)
		}

		// exit by design - usually a user only performs one extraction and then closes garbo
		os.Exit(0)
	}

	dialog := dialog.NewFolderOpen(dialogCallback, dialogWindow)
	location, _ := storage.ListerForURI(storage.NewFileURI(path.Dir(filePath)))
	dialog.SetLocation(location)
	dialog.Show()
	time.Sleep(100 * time.Millisecond)
	dialog.Resize(dialogSize)
}

func buildTree(filePath string) (fyne.CanvasObject, error) {
	treeData := make(map[string]UIFile)
	dirData := make(map[string][]UIFile)
	containers := make(map[*fyne.Container]*FileItem)

	err := walkArchive(filePath, treeData, dirData)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	childUIDsFunc := func(uid widget.TreeNodeID) (c []widget.TreeNodeID) {
		out := []string{}
		for _, f := range dirData[uid] {
			out = append(out, f.Name)
		}
		return out
	}
	isBranchFunc := func(uid widget.TreeNodeID) (ok bool) {
		children, ok := dirData[uid]
		return ok && len(children) > 0
	}
	createFunc := func(branch bool) (o fyne.CanvasObject) {
		nameLabel := widget.NewLabel("")
		sizeLabel := widget.NewLabel("")
		extractButton := widget.NewButton("Extract", func() {})
		fi := &FileItem{
			NameLabel:     nameLabel,
			SizeLabel:     sizeLabel,
			ExtractButton: extractButton,
		}
		rightAligned := fyne.NewContainerWithLayout(
			layout.NewHBoxLayout(),
			sizeLabel,
			extractButton,
		)
		c := fyne.NewContainerWithLayout(
			layout.NewBorderLayout(nil, nil, nameLabel, rightAligned),
			nameLabel,
			rightAligned,
		)
		containers[c] = fi
		return c
	}
	updateFunc := func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
		name := path.Base(treeData[uid].Name)
		fi := containers[node.(*fyne.Container)]
		fi.NameLabel.SetText(name)
		if !treeData[uid].IsDir {
			fi.SizeLabel.SetText(humanize.Bytes(uint64(treeData[uid].Size)))
		}
		fi.ExtractButton.OnTapped = func() {
			go extractFile(filePath, treeData[uid])
		}
	}
	t := widget.NewTree(childUIDsFunc, isBranchFunc, createFunc, updateFunc)
	t.OnSelected = func(uid widget.TreeNodeID) {
		t.OpenBranch(uid)
	}
	return t, nil
}

func buildOpenView(mainWindow fyne.Window) *fyne.Container {
	openArchiveButtonCallback := func() {
		dialogSize := fyne.NewSize(700, 500)

		dialogWindow := fyne.CurrentApp().NewWindow("Open archive...")
		dialogWindow.Resize(dialogSize)
		dialogWindow.CenterOnScreen()
		dialogWindow.Show()

		dialogCallback := func(file fyne.URIReadCloser, err error) {
			if err != nil || file == nil {
				return
			}
			filePath := strings.TrimPrefix(file.URI().String(), "file://")
			content, err := buildArchiveView(mainWindow, filePath)
			if err == nil {
				mainWindow.SetContent(content)
			}
			dialogWindow.Close()
		}

		dialog := dialog.NewFileOpen(dialogCallback, dialogWindow)
		dialog.Show()
		time.Sleep(300 * time.Millisecond)
		dialog.Resize(dialogSize)
	}
	openArchiveButton := widget.NewButton("Open archive...", openArchiveButtonCallback)
	openArchiveButton.Importance = widget.HighImportance

	return fyne.NewContainerWithLayout(
		layout.NewCenterLayout(),
		openArchiveButton,
	)
}

func buildArchiveView(mainWindow fyne.Window, filePath string) (*fyne.Container, error) {
	top := widget.NewLabel(fmt.Sprintf("Viewing archive: %s", filePath))
	center, err := buildTree(filePath)
	if err != nil {
		errorLabel := widget.NewLabel(err.Error())
		widget.ShowPopUp(errorLabel, mainWindow.Canvas())
		return nil, err
	}
	return fyne.NewContainerWithLayout(
		layout.NewBorderLayout(top, nil, nil, nil),
		top,
		center,
	), nil
}

func main() {
	a := app.New()
	mainWindow := a.NewWindow("garbo")

	var content fyne.CanvasObject
	if len(os.Args) == 2 {
		if os.Args[1] == "-version" || os.Args[1] == "--version" {
			fmt.Println("garbo")
			fmt.Println("Copyright (c) 2021 Thomas Kriechbaumer")
			fmt.Println("See https://github.com/Kriechi/garbo for more information.")
			os.Exit(0)
			return
		}
		c, err := buildArchiveView(mainWindow, os.Args[1])
		if err != nil {
			content = buildOpenView(mainWindow)
		} else {
			content = c
		}
	} else {
		content = buildOpenView(mainWindow)
	}

	mainWindow.SetContent(content)
	mainWindow.Resize(fyne.NewSize(400, 500))
	mainWindow.CenterOnScreen()
	mainWindow.ShowAndRun()
}
