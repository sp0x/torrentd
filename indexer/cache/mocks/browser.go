package mocks

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
	"io"
	"net/http"
	"net/url"
	"time"
)

type MockedBrowser struct {
	CanOpen bool
	state   *jar.State
}

// SetUserAgent sets the user agent.
func (b *MockedBrowser) SetUserAgent(_ string) {

}

//SetEncoding sets the encoding to use after a page is fetched
func (b *MockedBrowser) SetEncoding(_ string) {

}
func (b *MockedBrowser) SetRateLimit(_ int) {

}

// SetAttribute sets a browser instruction attribute.
func (b *MockedBrowser) SetAttribute(_ browser.Attribute, _ bool) {

}

func (b *MockedBrowser) RawBody() []byte {
	return nil
}

// SetAttributes is used to set all the browser attributes.
func (b *MockedBrowser) SetAttributes(_ browser.AttributeMap) {

}

// SetState sets the init browser state.
func (b *MockedBrowser) SetState(state *jar.State) {
	b.state = state
}

// State returns the browser state.
func (b *MockedBrowser) State() *jar.State {
	return nil
}

// SetBookmarksJar sets the bookmarks jar the browser uses.
func (b *MockedBrowser) SetBookmarksJar(_ jar.BookmarksJar) {

}

// BookmarksJar returns the bookmarks jar the browser uses.
func (b *MockedBrowser) BookmarksJar() jar.BookmarksJar {
	return nil
}

// SetCookieJar is used to set the cookie jar the browser uses.
func (b *MockedBrowser) SetCookieJar(_ http.CookieJar) {

}

// CookieJar returns the cookie jar the browser uses.
func (b *MockedBrowser) CookieJar() http.CookieJar {
	return nil
}

// SetHistoryJar is used to set the history jar the browser uses.
func (b *MockedBrowser) SetHistoryJar(_ jar.History) {

}

// HistoryJar returns the history jar the browser uses.
func (b *MockedBrowser) HistoryJar() jar.History {
	return nil
}

// SetHeadersJar sets the headers the browser sends with each request.
func (b *MockedBrowser) SetHeadersJar(_ http.Header) {

}

// SetTimeout sets the timeout for requests.
func (b *MockedBrowser) SetTimeout(_ time.Duration) {

}

// SetTransport sets the http library transport mechanism for each request.
func (b *MockedBrowser) SetTransport(http.RoundTripper) {

}

// AddRequestHeader adds a header the browser sends with each request.
func (b *MockedBrowser) AddRequestHeader(string, string) {

}

// Open requests the given URL using the GET method.
func (b *MockedBrowser) Open(string) error {
	if b.CanOpen {
		return nil
	} else {
		return errors.New("couldn't connect")
	}
}

// Open requests the given URL using the HEAD method.
func (b *MockedBrowser) Head(string) error {
	return nil
}

// OpenForm appends the data values to the given URL and sends a GET request.
func (b *MockedBrowser) OpenForm(string, url.Values) error {
	return nil
}

// OpenBookmark calls Get() with the URL for the bookmark with the given name.
func (b *MockedBrowser) OpenBookmark(string) error {
	return nil
}

// Post requests the given URL using the POST method.
func (b *MockedBrowser) Post(string, string, io.Reader) error {
	return nil
}

// PostForm requests the given URL using the POST method with the given data.
func (b *MockedBrowser) PostForm(string, url.Values) error {
	return nil
}

// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
func (b *MockedBrowser) PostMultipart(string, url.Values, browser.FileSet) error {
	return nil
}

// Back loads the previously requested page.
func (b *MockedBrowser) Back() bool {
	return true
}

// Reload duplicates the last successful request.
func (b *MockedBrowser) Reload() error {
	return nil
}

// Bookmark saves the page URL in the bookmarks with the given name.
func (b *MockedBrowser) Bookmark(string) error {
	return nil
}

// Click clicks on the page element matched by the given expression.
func (b *MockedBrowser) Click(string) error {
	return nil
}

// Form returns the form in the current page that matches the given expr.
func (b *MockedBrowser) Form(string) (browser.Submittable, error) {
	return nil, nil
}

// Forms returns an array of every form in the page.
func (b *MockedBrowser) Forms() []browser.Submittable {
	return nil
}

// Links returns an array of every link found in the page.
func (b *MockedBrowser) Links() []*browser.Link {
	return nil
}

// Images returns an array of every image found in the page.
func (b *MockedBrowser) Images() []*browser.Image {
	return nil
}

// Stylesheets returns an array of every stylesheet linked to the document.
func (b *MockedBrowser) Stylesheets() []*browser.Stylesheet {
	return nil
}

// Scripts returns an array of every script linked to the document.
func (b *MockedBrowser) Scripts() []*browser.Script {
	return nil
}

// SiteCookies returns the cookies for the current site.
func (b *MockedBrowser) SiteCookies() []*http.Cookie {
	return nil
}

// ResolveUrl returns an absolute URL for a possibly relative URL.
func (b *MockedBrowser) ResolveUrl(*url.URL) *url.URL {
	return nil
}

// ResolveStringUrl works just like ResolveUrl, but the argument and return value are strings.
func (b *MockedBrowser) ResolveStringUrl(string) (string, error) {
	return "", nil
}

// Download writes the contents of the document to the given writer.
func (b *MockedBrowser) Download(io.Writer) (int64, error) {
	return 0, nil
}

// Url returns the page URL as a string.
func (b *MockedBrowser) Url() *url.URL {
	return nil
}

// StatusCode returns the response status code.
func (b *MockedBrowser) StatusCode() int {
	if b.CanOpen {
		return http.StatusOK
	} else {
		return 502
	}
}

// Title returns the page title.
func (b *MockedBrowser) Title() string {
	return ""
}

// ResponseHeaders returns the page headers.
func (b *MockedBrowser) ResponseHeaders() http.Header {
	return nil
}

// Body returns the page body as a string of html.
func (b *MockedBrowser) Body() string {
	return ""
}

// Dom returns the inner *goquery.Selection.
func (b *MockedBrowser) Dom() *goquery.Selection {
	return b.state.Dom.Contents()
}

// Find returns the dom selections matching the given expression.
func (b *MockedBrowser) Find(string) *goquery.Selection {
	return nil
}

// Create a new Browser instance and inherit the configuration
// Read more: https://github.com/headzoo/surf/issues/23
func (b *MockedBrowser) NewTab() (bx *browser.Browser) {
	return nil
}
