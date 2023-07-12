package crud

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MediaFrontend struct {
	nested      ResourceFrontend
	unpoliced   Resource
	tmpFolder   string
	finalFolder string
	mimeTypes   map[string]string
	ffmpegPath  string
}

// FromMedia creates a new MediaFrontend
func FromMedia(r Resource, unpoliced Resource, tmpFolder, finalFolder string, mimeTypes map[string]string, ffmpegPath string) MediaFrontend {
	for _, ext := range mimeTypes {
		if !strings.HasPrefix(ext, ".") {
			panic("mimetype extensions must begin with `.`")
		}
	}
	return MediaFrontend{
		nested:      FromResource(r),
		unpoliced:   unpoliced,
		tmpFolder:   tmpFolder,
		finalFolder: finalFolder,
		mimeTypes:   mimeTypes,
		ffmpegPath:  ffmpegPath,
	}
}

// MediaFolder returns the root media folder
func (h MediaFrontend) MediaFolder() string {
	return h.finalFolder
}

// Get handler
func (h MediaFrontend) Get(r *http.Request) (io.ReadCloser, error) {
	return h.nested.Get(r)
}

type mediaResponse struct {
	ID       string `json:"id"`
	MediaURL string `json:"media_url"`
}

// Post handler
func (h MediaFrontend) Post(r *http.Request) (io.ReadCloser, error) {
	if r.Body == nil {
		return nil, ErrEmptyBody
	}
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return h.nested.Post(r)
	}
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}
	requestParams := make(map[string]string)
	requestParams["id"] = id
	idFolder := idFolder(id)
	escapeId := escapeId(id)
	var (
		fileExt string
		tmpPath string
	)
	// We must clean "tmpPath" variable if upload succeeds
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()
	// This closure will update tmpPath and fileExt above
	processPart := func(p *multipart.Part) error {
		defer exhaust(p)
		formName := p.FormName()
		if formName == "" {
			return ErrMultipartName
		}
		fileName := p.FileName()
		if fileName == "" {
			// This is a request parameter
			encodedContent, err := io.ReadAll(p)
			if err != nil {
				return err
			}
			content, err := url.PathUnescape(string(encodedContent))
			if err != nil {
				return err
			}
			requestParams[formName] = content
		} else {
			if tmpPath != "" {
				return ErrMultipartTooManyFiles
			}
			contentType := p.Header.Get("Content-Type")
			if contentType == "" {
				return ErrMultipartNeedsContentType
			}
			fileExt, err = h.checkMimeType(contentType)
			if err != nil {
				return err
			}
			tmpPath, err = saveTmpFile(h.tmpFolder, escapeId, p)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// so that we can defer exhaust(p)
		if err := processPart(p); err != nil {
			return nil, err
		}
	}
	if tmpPath == "" {
		return nil, ErrMultipartNoFile
	}
	// Try to transcode AVI files, so that they can be played in the browser
	if h.ffmpegPath != "" && strings.HasSuffix(strings.ToLower(fileExt), ".avi") {
		transcode := func() {
			// try to convert to mp4 using ffmpeg
			// this is a best effort, so we ignore errors
			// and just keep the original file
			outPath := strings.TrimSuffix(tmpPath, fileExt) + ".mp4"
			// See https://superuser.com/questions/710008/how-to-get-rid-of-ffmpeg-pts-has-no-value-error
			// for an explanation of -fflags
			cmd := exec.CommandContext(r.Context(), h.ffmpegPath, "-fflags", "+genpts", "-i", tmpPath, "-c:v", "copy", "-c:a", "copy", "-y", outPath)
			if err := cmd.Run(); err != nil {
				log.Printf("ffmpeg failed: %v", err)
				return
			}
			log.Printf("ffmpeg transcoded %s to %s", tmpPath, outPath)
			// ffmpeg succeeded, so we can delete the original file
			os.Remove(tmpPath)
			// and update tmpPath and fileExt
			tmpPath = outPath
			fileExt = ".mp4"
		}
		transcode()
	}
	mediaURL, err := h.commitTmpFile(r.Context(), id, idFolder, escapeId, fileExt, tmpPath)
	if err != nil {
		return nil, err
	}
	// Best effort: write a "meta" file for each upload, with the request parameters
	requestParams["media_url"] = mediaURL
	metaFile := h.metaFile(idFolder, escapeId)
	if meta, err := os.Create(metaFile); err == nil {
		enc := json.NewEncoder(meta)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		enc.Encode(requestParams)
		meta.Close()
	}
	// Return the id and media_url to whomever is interested
	response := mediaResponse{
		ID:       id,
		MediaURL: mediaURL,
	}
	result, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(result)), nil
}

// Put handler
func (h MediaFrontend) Put(r *http.Request) error {
	return h.nested.Put(r)
}

// Delete handler
func (h MediaFrontend) Delete(r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return ErrMissingResourceId
	}
	escapeId := escapeId(id)
	idFolder := idFolder(id)
	if r.URL.Query().Get("mediaOnly") == "true" {
		// delete only media files
		err := h.removePrevFiles(idFolder, escapeId)
		if err != nil {
			os.Remove(h.metaFile(idFolder, escapeId))
		}
		return err
	}
	err := h.unpoliced.Delete(r.Context(), id)
	if err == nil {
		// Remove prev files only if we deleted the resource
		err = h.removePrevFiles(idFolder, escapeId)
		if err != nil {
			os.Remove(h.metaFile(idFolder, escapeId))
		}
	}
	return err
}

func (h MediaFrontend) checkMimeType(contentType string) (string, error) {
	for mediaType, ext := range h.mimeTypes {
		if strings.HasPrefix(contentType, mediaType) {
			return ext, nil
		}
	}
	return "", ErrMimeNotSupported
}

// saveFile saves the input stream as a file
func (h MediaFrontend) commitTmpFile(ctx context.Context, id, idFolder, escapeId, ext, tmpPath string) (mediaURL string, err error) {
	// make storage folder
	if err = os.MkdirAll(idFolder, 0755); err != nil {
		return "", err
	}
	// Find existing files
	var matches []string
	matches, err = h.prevFiles(idFolder, escapeId)
	if err != nil {
		return "", err
	}
	// rename previous files
	renamed := make(map[string]string)
	defer func() {
		if err == nil {
			// remove old files
			for _, newName := range renamed {
				os.Remove(newName)
			}
		} else {
			// try to restore prev files
			for oldName, newName := range renamed {
				os.Rename(newName, oldName)
			}
		}
	}()
	for _, match := range matches {
		newName := match + ".old"
		if err = os.Rename(match, match+".old"); err != nil {
			return "", err
		}
		renamed[match] = newName
	}
	// make sure the final folder exists
	if err = os.MkdirAll(filepath.Join(h.finalFolder, idFolder), 0755); err != nil {
		return "", err
	}
	// move to final location. Notice: `ext` already includes the dot.
	finalName := fmt.Sprintf("%s%s", escapeId, ext)
	finalPath := filepath.Join(h.finalFolder, idFolder, finalName)
	if err = os.Rename(tmpPath, finalPath); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			os.Remove(finalPath)
		}
	}()
	// Update resource's `mediaURL` attrib with the new file
	mediaURL = strings.Join([]string{idFolder, finalName}, "/")
	params := map[string]string{
		"media_url": mediaURL,
	}
	var data []byte
	data, err = json.Marshal(params)
	if err != nil {
		return "", err
	}
	err = h.unpoliced.Put(ctx, id, bytes.NewReader(data))
	return mediaURL, err
}

// idFolder builds path from ID and extension
func idFolder(id string) string {
	hash := fnv.New64a()
	hash.Write([]byte(id))
	return fmt.Sprintf("%03d", hash.Sum64()&0x0FF)
}

// escape path to avoid directory traversal
func escapeId(id string) string {
	return base64.URLEncoding.EncodeToString([]byte(id))
}

// prevFiles finds any prevoius files associated to this id
func (h MediaFrontend) prevFiles(idFolder, escapeId string) ([]string, error) {
	idGlob := fmt.Sprintf("%s.*", escapeId)
	return filepath.Glob(filepath.Join(h.finalFolder, idFolder, idGlob))
}

// saveFile saves the input stream as a file
func saveTmpFile(tmpFolder, escapeId string, p io.ReadCloser) (tmpPath string, err error) {
	// save to temporary file
	tmpPath = filepath.Join(tmpFolder, escapeId)
	var tmpFile *os.File
	tmpFile, err = os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()
	defer tmpFile.Close()
	_, err = io.Copy(tmpFile, p)
	if err != nil {
		return "", err
	}
	return tmpPath, nil
}

func (h MediaFrontend) metaFile(idFolder, escapeId string) string {
	return filepath.Join(h.finalFolder, idFolder, fmt.Sprintf("%s.meta", escapeId))
}

// remove files associated to this id
func (h MediaFrontend) removePrevFiles(idFolder, escapeId string) error {
	matches, err := h.prevFiles(idFolder, escapeId)
	if err != nil {
		return err
	}
	for _, match := range matches {
		err = errors.Join(os.Remove(match))
	}
	return err
}
