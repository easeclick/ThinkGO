package thinkgo

import (
	"bytes"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ViewEngine renders templates using Go's html/template.
// Similar to ThinkPHP's template/view system.
type ViewEngine struct {
	mu         sync.RWMutex
	templates  *template.Template
	dir        string
	extension  string
	cached     bool
	funcMap    template.FuncMap
}

// NewViewEngine creates a new template engine.
// dir is the template directory (e.g., "view/").
func NewViewEngine(dir string) *ViewEngine {
	v := &ViewEngine{
		dir:       dir,
		extension: ".html",
		cached:    true,
		funcMap:   make(template.FuncMap),
	}

	// Add built-in functions
	v.funcMap["safe"] = func(s string) template.HTML {
		return template.HTML(s)
	}

	v.loadTemplates()
	return v
}

// SetExtension sets the template file extension.
func (v *ViewEngine) SetExtension(ext string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.extension = ext
	v.loadTemplates()
}

// SetCached enables/disables template caching.
func (v *ViewEngine) SetCached(cached bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cached = cached
}

// AddFunc adds a template function.
func (v *ViewEngine) AddFunc(name string, fn any) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.funcMap[name] = fn
}

// Render renders a template with the given data.
func (v *ViewEngine) Render(name string, data any) (string, error) {
	v.mu.RLock()
	tmpl := v.templates
	cached := v.cached
	v.mu.RUnlock()

	// Reload in dev mode
	if !cached {
		v.mu.Lock()
		v.loadTemplates()
		tmpl = v.templates
		v.mu.Unlock()
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderToWriter renders a template directly to a writer.
func (v *ViewEngine) RenderToWriter(name string, data any, ctx *Context) error {
	v.mu.RLock()
	tmpl := v.templates
	cached := v.cached
	v.mu.RUnlock()

	if !cached {
		v.mu.Lock()
		v.loadTemplates()
		tmpl = v.templates
		v.mu.Unlock()
	}

	ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(ctx.Writer, name, data)
}

// loadTemplates loads all template files from the directory.
func (v *ViewEngine) loadTemplates() {
	tmpl := template.New("").Funcs(v.funcMap)

	// Check if dir exists
	if _, err := os.Stat(v.dir); os.IsNotExist(err) {
		v.templates = tmpl
		return
	}

	filepath.WalkDir(v.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, v.extension) {
			return nil
		}

		// Get template name relative to dir
		relPath, _ := filepath.Rel(v.dir, path)
		name := strings.TrimSuffix(relPath, v.extension)
		name = strings.ReplaceAll(name, string(filepath.Separator), "/")

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		template.Must(tmpl.New(name).Parse(string(content)))
		return nil
	})

	v.templates = tmpl
}
