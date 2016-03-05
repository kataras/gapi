package iris

import (
	"net/http"
	"strings"
)

// Route contains its middleware, handler, pattern , it's path string, http methods and a template cache
// Used to determinate which handler on which path must call
// Used on router.go
type Route struct {
	//GET, POST, PUT, DELETE, CONNECT, HEAD, PATCH, OPTIONS, TRACE bool //tried with []string, very slow, tried with map[string]bool gives 10k executions but +20k bytes, with this approact we have to code more but only 1k byte to space and make it 2.2 times faster than before!
	//Middleware
	MiddlewareSupporter
	//mu            sync.RWMutex

	pathPrefix string // this is to make a little faster the match, before regexp Match runs, it's the path before the first path parameter :
	//the pathPrefix is with the last / if parameters exists.
	parts []string //stores the string path AFTER the pathPrefix, without the pathPrefix. no need to that but no problem also.
	//if parts != nil means that this route has no params
	fullpath string // need only on parameters.Params(...)
	//fullparts   []string
	handler     Handler
	templates   *TemplateCache //this is passed to the Renderer
	httpErrors  *HTTPErrors    //the only need of this is to pass into the Context, in order to  developer get the ability to perfom emit errors (eg NotFound) directly from context
	hasWildcard bool
	isReady     bool
}

// newRoute creates, from a path string, handler and optional http methods and returns a new route pointer
func newRoute(registedPath string, handler Handler) *Route {
	r := &Route{handler: handler}

	hasPathParameters := false
	firstPathParamIndex := strings.IndexByte(registedPath, ParameterStartByte)
	if firstPathParamIndex != -1 {
		r.pathPrefix = registedPath[:firstPathParamIndex]
		hasPathParameters = true

		if strings.HasSuffix(registedPath, MatchEverything) {
			r.hasWildcard = true
		}

	} else {
		//check for only for* , here no path parameters registed.
		firstPathParamIndex = strings.IndexByte(registedPath, MatchEverythingByte)

		if firstPathParamIndex != -1 {
			if firstPathParamIndex <= 1 { // set to '*' to pathPrefix if no prefix exists except the slash / if any [Get("/*",..) or Get("*",...]
				//has no prefix just *
				r.pathPrefix = MatchEverything
				r.hasWildcard = true
			} else { //if firstPathParamIndex == len(registedPath)-1 { // it's the last
				//has some prefix and sufix of *
				r.pathPrefix = registedPath[:firstPathParamIndex] //+1
				r.hasWildcard = true
			}

		} else {
			//else no path parameter or match everything symbol so use the whole path as prefix it will be faster at the check for static routes too!
			r.pathPrefix = registedPath
		}

	}

	if hasPathParameters || r.hasWildcard {
		r.parts = strings.Split(registedPath[len(r.pathPrefix):], "/")
		//r.fullparts = strings.Split(registedPath[1:], "/")
	}
	r.fullpath = registedPath //we need this only to take Params so set it if has path parameters.

	return r
}

// containsMethod determinates if this route contains a http method
// match determinates if this route match with the request, returns bool as first value and PathParameters as second value, if any
func (r *Route) match(urlPath string) bool {
	/* we do it on r.pathPrefix == urlPath if r.parts == nil { //if this route doesn't support params
		return r.fullpath == urlPath
	}*/
	if r.pathPrefix == MatchEverything {
		return true
	}
	if r.pathPrefix == urlPath {
		//it's route without path parameters or * symbol, and if the request url has prefix of it  and it's the same as the whole preffix which is the path itself returns true without checking for regexp pattern
		//so it's just a path without named parameters
		return true
	}

	//kapws kai to sufix na vlepw an den einai parameter, an einai idio kai meta na sunexizei sto path parameters.

	s := urlPath[len(r.pathPrefix):]
	if s[0] == '/' { //it's whrong by the way, the after pathPrefix can't be start with /, because the route'spathPrefix ends with '/'
		return false
	}
	urlPathLen := len(s)
	partsLen := len(r.parts)
	lastStartPart := 1
	partIndex := 0
	urlPathMin := urlPathLen - 1
	var reqPart string
	for i := 0; i < urlPathLen; i++ {
		if i == urlPathMin || s[i] == '/' { //to prwto einai to slash panta giauto to pernaw  //an exoume part i eimaste sto telos
			isTheLastPart := partIndex == partsLen-1
			if r.hasWildcard && isTheLastPart {
				//means we are at the last part
				//if r.parts[partsLen-1][0] == '*' {

				return true //it ends with /home/test/*
				//}
			}
			if partsLen <= partIndex {
				return false
			}
			iplus := i + 1
			if r.parts[partIndex][0] == ':' {
				//if isTheLastPart && iplus == urlPathMin {
				//works but no so much difference at the perfomance.
				//return true //it ends with /home/test/:name
				//}
				partIndex++
				lastStartPart = iplus
				continue
			}
			if i == urlPathMin { //last part
				reqPart = s[lastStartPart:iplus]
			} else {
				reqPart = s[lastStartPart:i]
			}

			lastStartPart = iplus

			if r.parts[partIndex] != reqPart {
				return false
			}
			partIndex++
		}

	}

	if partIndex < partsLen {
		return false
	}
	//go func(url string, route *Route) {
	//exists := route.cache[url] == true
	//if !exists {
	//	if len(route.cache) > 4 {
	//		route.cache = append(route.cache[:0], url)
	//	} else {
	//		println("cache ", url)
	//route.cache = append(route.cache, url)
	//	}

	//}

	//}(urlPath, r)

	return true

}

func getMiddle(str string) string {
	if (len(str) % 2) == 0 {
		// Even length

		if len(str) > 2 {

			return str[len(str)/2-1 : len(str)/2+1]

		}

	}

	// Odd length
	return str[len(str)/2 : len(str)/2+1]
}

func strEqual(s string, s2 string) bool {
	slen := len(s)
	if len(s2) != slen {
		return false
	}

	if s[0] != s2[0] || s[slen-1] != s2[slen-1] {
		return false
	}

	if getMiddle(s) != getMiddle(s2) {
		return false
	}

	return s == s2
}

// Template creates (if not exists) and returns the template cache for this route
func (r *Route) Template() *TemplateCache {
	if r.templates == nil {
		r.templates = NewTemplateCache()
	}
	return r.templates
}

// prepare prepares the route's handler , places it to the last middleware , handler acts like a middleware too.
// Runs once before the first ServeHTTP
// MUST REMOVE IT SOME DAY AND MAKE MIDDLEWARE MORE LIGHTER
func (r *Route) prepare() {
	//r.mu.Lock()
	//look why on router ->HandleFunc defer r.mu.Unlock()
	//but wait... do we need locking here?

	convertedMiddleware := MiddlewareHandlerFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		r.handler.run(r, res, req)
		next(res, req)
	})

	r.Use(convertedMiddleware)
	r.isReady = true

}

// ServeHTTP serves this route and it's middleware
func (r *Route) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if r.middlewareHandlers != nil {
		if !r.isReady {
			r.prepare()
		}
		r.middleware.ServeHTTP(res, req)
	} else {
		r.handler.run(r, res, req)
	}
}
