package mdo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/web/returnerror"
)

func Mware(isMDoReq func(*http.Request) bool, root, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isMDoReq(r) {
			var err error
			mDoReqs := map[string]*mDoReq{}
			argsStr := r.URL.Query().Get("args")
			argsBytes := []byte(argsStr)
			if argsStr == "" {
				argsBytes, err = ioutil.ReadAll(r.Body)
				PanicOn(err)
			}
			err = json.Unmarshal(argsBytes, &mDoReqs)
			returnerror.If(err != nil, http.StatusBadRequest, "error unmarshalling json: %s", err)
			fullMDoResponse := map[string]*mDoResponse{}
			fullMDoResponseMtx := &sync.Mutex{}
			does := make([]Fn, 0, len(mDoReqs))
			for key := range mDoReqs {
				does = append(does, func(key string, mdoReq *mDoReq) func() {
					return func() {
						argsBytes, err := json.Marshal(mdoReq.Args)
						PanicOn(err)
						subReq, err := http.NewRequest(http.MethodPut, mdoReq.Path+"?isSubMDo=true", bytes.NewReader(argsBytes))
						PanicOn(err)
						PanicIf(isMDoReq(subReq), "can't have mdo request inside an mdo request")
						for _, c := range r.Cookies() {
							subReq.AddCookie(c)
						}
						for name := range r.Header {
							subReq.Header.Add(name, r.Header.Get(name))
						}
						subResp := &mDoResponseWriter{header: http.Header{}, body: bytes.NewBuffer(make([]byte, 0, 10000))}
						root(subResp, subReq)
						fullMDoResponseMtx.Lock()
						defer fullMDoResponseMtx.Unlock()
						fullMDoResponse[key] = &mDoResponse{
							headers: mdoReq.headers,
							Status:  subResp.status,
							Header:  subResp.header,
							Body:    subResp.body.Bytes(),
						}
					}
				}(key, mDoReqs[key]))
			}
			PanicOn(GoGroup(does...))
			json.WriteHttpOk(w, fullMDoResponse)
		} else {
			next(w, r)
		}
	}
}

func IsSubMDo(r *http.Request) bool {
	return r.URL.Query().Get("isSubMDo") == "true"
}

type mDoReq struct {
	headers bool                   `json:"headers"`
	Path    string                 `json:"path"`
	Args    map[string]interface{} `json:"args"`
}

type mDoResponseWriter struct {
	status int
	header http.Header
	body   *bytes.Buffer
}

func (r *mDoResponseWriter) Header() http.Header {
	return r.header
}

func (r *mDoResponseWriter) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *mDoResponseWriter) WriteHeader(code int) {
	r.status = code
}

type mDoResponse struct {
	headers bool
	Status  int         `json:"status"`
	Header  http.Header `json:"header"`
	Body    []byte      `json:"body"`
}

func (r *mDoResponse) MarshalJSON() ([]byte, error) {
	if r.headers {
		h, err := json.Marshal(r.Header)
		PanicOn(err)
		return []byte(Sprintf(`{"status":%d,"header":%s,"body":%s}`, r.Status, h, r.Body)), nil
	} else {
		return []byte(Sprintf(`{"status":%d,"body":%s}`, r.Status, r.Body)), nil
	}
}
