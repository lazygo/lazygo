package engine

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

// WrapHandler wraps `http.Handler` into `HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx.SetRequest(r)
				ctx.SetResponseWriter(NewResponseWriter(w))
				err = next(ctx)
			})).ServeHTTP(ctx.ResponseWriter(), ctx.Request())
			return
		}
	}
}

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

// AssetHandler returns an http.Handler that will serve files from
// the Assets embed.FS. When locating a file, it will strip the given
// prefix from the request and prepend the root to the filesystem.
func AssetHandler(prefix string, assets embed.FS, root string) HandlerFunc {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join(root, name)

		// If we can't find the asset, fs can handle the error
		file, err := assets.Open(assetPath)
		if err != nil {
			return nil, err
		}

		// Otherwise assume this is a legitimate request routed correctly
		return file, err
	})

	return WrapHandler(http.StripPrefix(prefix, http.FileServer(http.FS(handler))))
}
