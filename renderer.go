package iris

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

const (
	// DefaultCharset represents the default charset for content headers
	DefaultCharset = "UTF-8"
	// ContentType represents the header["Content-Type"]
	ContentType = "Content-Type"
	// ContentLength represents the header["Content-Length"]
	ContentLength = "Content-Length"
	// ContentHTML is the  string of text/html response headers
	ContentHTML = "text/html" + "; " + DefaultCharset
	// ContentJSON is the  string of application/json response headers
	ContentJSON = "application/json" + "; " + DefaultCharset
	// ContentJSONP is the  string of application/javascript response headers
	ContentJSONP = "application/javascript"
	// ContentBINARY is the  string of "application/octet-stream response headers
	ContentBINARY = "application/octet-stream"
	// ContentTEXT is the  string of text/plain response headers
	ContentTEXT = "text/plain" + "; " + DefaultCharset
	// ContentXML is the  string of text/xml response headers
	ContentXML = "text/xml" + "; " + DefaultCharset
)

var rendererType reflect.Type

// Renderer is the container of the template cache which developer creates for EACH route
type Renderer struct {
	//Only one TemplateCache per app/router/iris instance.
	//and for now Renderer writer content-type  doesn't checks for methods (get,post...)
	templateCache  *TemplateCache
	responseWriter http.ResponseWriter
}

// newRenderer creates and returns a new Renderer pointer
// Used at route.run
func newRenderer(writer http.ResponseWriter) *Renderer {
	return &Renderer{responseWriter: writer}
}

func (r *Renderer) check() error {
	if r.templateCache == nil {
		return errors.New("iris:Error on Renderer : No Template Cache was created yet, please refer to docs at github.com/kataras/iris")
	}
	return nil
}

// RenderFile renders a file by its path and a context passed to the function
func (r *Renderer) RenderFile(file string, pageContext interface{}) error {
	err := r.check()
	if err != nil {
		return err
	}

	return r.templateCache.ExecuteTemplate(r.responseWriter, file, pageContext)

}

// Render renders the template file html which is already registed to the template cache, with it's pageContext passed to the function
func (r *Renderer) Render(pageContext interface{}) error {
	err := r.check()
	if err != nil {
		return err
	}
	return r.templateCache.Execute(r.responseWriter, pageContext)

}

// WriteHTML writes html string with a http status
///TODO or I will think to pass an interface on handlers as second parameter near to the Context, with developer's custom Renderer package .. I will think about it.
func (r *Renderer) WriteHTML(httpStatus int, htmlContents string) {
	r.responseWriter.Header().Set(ContentType, ContentHTML)
	r.responseWriter.WriteHeader(httpStatus)
	io.WriteString(r.responseWriter, htmlContents)
}

//HTML calls the WriteHTML with the 200 http status ok
func (r *Renderer) HTML(htmlContents string) {
	r.WriteHTML(http.StatusOK, htmlContents)
}

// WriteData writes binary data with a http status
func (r *Renderer) WriteData(httpStatus int, binaryData []byte) {
	r.responseWriter.Header().Set(ContentType, ContentBINARY)
	r.responseWriter.Header().Set(ContentLength, strconv.Itoa(len(binaryData)))
	r.responseWriter.WriteHeader(httpStatus)
	r.responseWriter.Write(binaryData)
}

//Data calls the WriteData with the 200 http status ok
func (r *Renderer) Data(binaryData []byte) {
	r.WriteData(http.StatusOK, binaryData)
}

// WriteText writes text with a http status
func (r *Renderer) WriteText(httpStatus int, text string) {
	r.responseWriter.Header().Set(ContentType, ContentTEXT)
	r.responseWriter.WriteHeader(httpStatus)
	io.WriteString(r.responseWriter, text)
}

//Text calls the WriteText with the 200 http status ok
func (r *Renderer) Text(text string) {
	r.WriteText(http.StatusOK, text)
}

// WriteJSON writes which is converted from struct(s) with a http status which they passed to the function via parameters
func (r *Renderer) WriteJSON(httpStatus int, jsonStructs ...interface{}) error {

	//	return json.NewEncoder(r.responseWriter).Encode(obj)
	var _json string
	for _, jsonStruct := range jsonStructs {
		theJSON, err := json.MarshalIndent(jsonStruct, "", "  ")
		if err != nil {
			//http.Error(r.responseWriter, err.Error(), http.StatusInternalServerError)
			return err
		}
		_json += string(theJSON) + "\n"
	}

	//keep in mind http.DetectContentType(data)
	//also we don't check if already header's content-type exists.
	r.responseWriter.Header().Set(ContentType, ContentJSON)
	r.responseWriter.WriteHeader(httpStatus)
	io.WriteString(r.responseWriter, _json)

	return nil
}

//JSON calls the WriteJSON with the 200 http status ok
func (r *Renderer) JSON(jsonStructs ...interface{}) error {
	return r.WriteJSON(http.StatusOK, jsonStructs)
}

// WriteJSONP writes jsonp by  converted from struct(s) with a http status which they passed to the function via parameters
///TODO: NOT READY YET
func (r *Renderer) WriteJSONP(httpStatus int, obj interface{}) {
	r.responseWriter.Header().Set(ContentType, ContentJSONP)
	r.responseWriter.WriteHeader(httpStatus)
}

//JSONP calls the WriteJSONP with the 200 http status ok
func (r *Renderer) JSONP(obj interface{}) {
	r.WriteJSONP(http.StatusOK, obj)
}

// WriteXML writes xml which is converted from struct(s) with a http status which they passed to the function via parameters
func (r *Renderer) WriteXML(httpStatus int, xmlStructs ...interface{}) error {

	var _xmlDoc string
	for _, xmlStruct := range xmlStructs {
		theDoc, err := xml.MarshalIndent(xmlStruct, "", "  ")
		if err != nil {
			return err
		}
		_xmlDoc += string(theDoc) + "\n"
	}
	r.responseWriter.Header().Set(ContentType, ContentXML)
	r.responseWriter.WriteHeader(httpStatus)
	io.WriteString(r.responseWriter, xml.Header+_xmlDoc)
	return nil
}

//XML calls the WriteXML with the 200 http status ok
func (r *Renderer) XML(xmlStructs ...interface{}) error {
	return r.WriteXML(http.StatusOK, xmlStructs)
}
