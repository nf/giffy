/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command giffy reads all the JPEG and PNG files from the current directory
// and writes them to an animated GIF as "out.gif".
// It has no configuration knobs whatsoever.
package main

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "image/jpeg"
	_ "image/png"
)

func main() {
	fs, err := dirFiles(".")
	if err != nil {
		log.Fatal("error reading current directory:", err)
	}
	var ms []*image.Paletted
	for i, n := range fs {
		log.Printf("Reading %v [%d/%d]\n", n, i+1, len(fs))
		m, err := readImage(n)
		if err != nil {
			log.Fatalf("error reading image: %v: %v", n, err)
		}
		r := m.Bounds()
		pm := image.NewPaletted(r, palette.Plan9)
		draw.FloydSteinberg.Draw(pm, r, m, image.ZP)
		ms = append(ms, pm)
	}
	ds := make([]int, len(ms))
	for i := range ds {
		ds[i] = 10
	}
	const out = "out.gif"
	log.Println("Generating", out)
	f, err := os.Create(out)
	if err != nil {
		log.Fatalf("error creating %v: %v", out, err)
	}
	err = gif.EncodeAll(f, &gif.GIF{Image: ms, Delay: ds, LoopCount: -1})
	if err != nil {
		log.Fatalf("error writing %v: %v", out, err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("error closing %v: %v", out, err)
	}
}

var validExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

func dirFiles(dir string) (names []string, err error) {
	fs, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, fi := range fs {
		n := fi.Name()
		if !validExtensions[filepath.Ext(n)] {
			continue
		}
		names = append(names, n)
	}
	sort.Sort(filenames(names))
	return
}

type filenames []string

func (s filenames) Len() int      { return len(s) }
func (s filenames) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s filenames) Less(i, j int) bool {
	if filepath.Ext(s[i]) == filepath.Ext(s[j]) {
		a, b := stripExt(s[i]), stripExt(s[j])
		if (strings.HasPrefix(a, b) || strings.HasPrefix(b, a)) &&
			strings.Contains(a, "#") != strings.Contains(b, "#") {
			return strings.Contains(b, "#")
		}
	}
	return s[i] < s[j]
}

func stripExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func readImage(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	return m, err
}
