package imports

// packages represents a list of common packages that can be imported by several
// templates.
var packages = map[string]*Import{
	"context": {
		Name: "context",
	},
	"errors": {
		Name: "errors",
	},
	"fmt": {
		Name: "fmt",
	},
	"json": {
		Name: "encoding/json",
	},
	"math/rand": {
		Name: "math/rand",
	},
	"reflect": {
		Name: "reflect",
	},
	"regex": {
		Name: "regexp",
	},
	"strings": {
		Name: "strings",
	},
	"time": {
		Name: "time",
	},
	"prototimestamp": {
		Name:  "google.golang.org/protobuf/types/known/timestamppb",
		Alias: "ts",
	},
	"protostruct": {
		Name: "google.golang.org/protobuf/types/known/structpb",
	},
	"fasthttp": {
		Name: "github.com/valyala/fasthttp",
	},
	"fasthttp-router": {
		Name: "github.com/fasthttp/router",
	},
	"validation": {
		Name: "github.com/go-ozzo/ozzo-validation/v4",
	},
}
