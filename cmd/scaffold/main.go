package main

import (
    "bytes"
    "errors"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "text/template"
)

type Field struct {
    Name      string
    GoName    string
    ProtoType string
    SQLType   string
    JSONName  string
    IsID      bool
    IsTS      bool
}

type Model struct {
    Module        string
    Name          string
    NameLower     string
    Table         string
    Service       string
    ProtoPackage  string
    GoPackagePath string
    GoPkgName     string
    Fields        []Field
}

func main() {
    var name string
    var fields string
    flag.StringVar(&name, "name", "", "Entity name in PascalCase, e.g. User")
    flag.StringVar(&fields, "fields", "", `Fields, e.g. "name:string email:string age:int"`)
    flag.Parse()

    if strings.TrimSpace(name) == "" {
        exitErr(errors.New("-name is required (e.g. -name User)"))
    }
    if strings.TrimSpace(fields) == "" {
        exitErr(errors.New(`-fields is required (e.g. -fields "name:string email:string age:int")`))
    }

    module, err := readModulePath("go.mod")
    if err != nil {
        exitErr(err)
    }

    m, err := buildModel(module, name, fields)
    if err != nil {
        exitErr(err)
    }

    if err := writeFromTemplate("proto", filepath.Join("proto", m.NameLower, "v1", m.NameLower+".proto"), protoTmpl, m); err != nil {
        exitErr(err)
    }
    if err := writeFromTemplate("usecase", filepath.Join("internal", "usecase", m.NameLower+"_usecase.go"), usecaseTmpl, m); err != nil {
        exitErr(err)
    }
    if err := writeFromTemplate("handler", filepath.Join("internal", "adapter", "grpc", m.NameLower+"_handler.go"), handlerTmpl, m); err != nil {
        exitErr(err)
    }
    if err := writeFromTemplate("repo-mem", filepath.Join("internal", "adapter", "repository", "memory", m.NameLower+"_repository.go"), repoMemoryTmpl, m); err != nil {
        exitErr(err)
    }
    if err := writeFromTemplate("repo-mysql", filepath.Join("internal", "adapter", "repository", "mysql", m.NameLower+"_repository.go"), repoMySQLTmpl, m); err != nil {
        exitErr(err)
    }

    if err := ensureSchemaSQL(filepath.Join("db", "schema.sql"), schemaTmpl, m); err != nil {
        exitErr(err)
    }
    if err := patchServerMain(filepath.Join("cmd", "server", "main.go"), m); err != nil {
        exitErr(err)
    }

    fmt.Printf("scaffolded: %s (fields: %d)\n", m.Name, len(m.Fields))
}

func exitErr(err error) {
    fmt.Fprintln(os.Stderr, "ERROR:", err)
    os.Exit(1)
}

func readModulePath(goModPath string) (string, error) {
    b, err := os.ReadFile(goModPath)
    if err != nil {
        return "", err
    }
    re := regexp.MustCompile(`(?m)^\s*module\s+(\S+)\s*$`)
    m := re.FindStringSubmatch(string(b))
    if len(m) != 2 {
        return "", fmt.Errorf("failed to parse module path from %s", goModPath)
    }
    return m[1], nil
}

func buildModel(module, name, fields string) (Model, error) {
    if !isPascal(name) {
        return Model{}, fmt.Errorf("name must be PascalCase (e.g. User, BlogPost). got: %q", name)
    }
    nl := strings.ToLower(name[:1]) + name[1:]
    table := pluralizeSnake(toSnake(nl))

    fp, err := parseFields(fields)
    if err != nil {
        return Model{}, err
    }

    m := Model{
        Module:        module,
        Name:          name,
        NameLower:     nl,
        Table:         table,
        Service:       name + "Service",
        ProtoPackage:  toSnake(nl) + ".v1",
        GoPkgName:     toSnake(nl) + "v1",
        GoPackagePath: fmt.Sprintf("%s/gen/%s/v1;%sv1", module, toSnake(nl), toSnake(nl)),
        Fields:        fp,
    }
    return m, nil
}

func parseFields(s string) ([]Field, error) {
    parts := strings.Fields(s)
    if len(parts) == 0 {
        return nil, errors.New("no fields provided")
    }
    var out []Field
    seen := map[string]bool{}
    for _, p := range parts {
        kv := strings.SplitN(p, ":", 2)
        if len(kv) != 2 {
            return nil, fmt.Errorf("invalid field spec: %q (expected name:type)", p)
        }
        name := strings.TrimSpace(kv[0])
        typ := strings.TrimSpace(kv[1])
        if name == "" || typ == "" {
            return nil, fmt.Errorf("invalid field spec: %q", p)
        }
        if !isLowerIdent(name) {
            return nil, fmt.Errorf("field name must be lower_snake/camel (lower start). got: %q", name)
        }
        if seen[name] {
            return nil, fmt.Errorf("duplicate field: %s", name)
        }
        seen[name] = true

        protoType, sqlType, err := mapType(typ)
        if err != nil {
            return nil, fmt.Errorf("field %s: %w", name, err)
        }
        out = append(out, Field{
            Name:      name,
            GoName:    toPascal(name),
            ProtoType: protoType,
            SQLType:   sqlType,
            JSONName:  toSnake(name),
        })
    }
    return out, nil
}

func mapType(t string) (protoType, sqlType string, err error) {
    switch strings.ToLower(t) {
    case "string":
        return "string", "VARCHAR(255)", nil
    case "text":
        return "string", "TEXT", nil
    case "int", "int32":
        return "int32", "INT", nil
    case "int64":
        return "int64", "BIGINT", nil
    case "bool":
        return "bool", "TINYINT(1)", nil
    default:
        return "", "", fmt.Errorf("unknown type %q (supported: string,text,int,int32,int64,bool)", t)
    }
}

func writeFromTemplate(label, path, tmpl string, m Model) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    if _, err := os.Stat(path); err == nil {
        return fmt.Errorf("%s already exists: %s", label, path)
    }
    t, err := template.New(label).Funcs(template.FuncMap{
        "inc": func(i int) int { return i + 1 },
        "add2": func(i int) int { return i + 2 },
    }).Parse(tmpl)
    if err != nil {
        return err
    }
    var buf bytes.Buffer
    if err := t.Execute(&buf, m); err != nil {
        return err
    }
    return os.WriteFile(path, buf.Bytes(), 0o644)
}

func ensureSchemaSQL(path, tmpl string, m Model) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    b, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            if err := os.WriteFile(path, []byte("-- schema.sql\n"), 0o644); err != nil {
                return err
            }
            b = []byte("-- schema.sql\n")
        } else {
            return err
        }
    }
    content := string(b)
    createSig := fmt.Sprintf("CREATE TABLE %s", m.Table)
    if strings.Contains(content, createSig) {
        return nil
    }
    t, err := template.New("schema").Parse(tmpl)
    if err != nil {
        return err
    }
    var buf bytes.Buffer
    if err := t.Execute(&buf, m); err != nil {
        return err
    }
    if !strings.HasSuffix(content, "\n") {
        content += "\n"
    }
    content += "\n" + buf.String() + "\n"
    return os.WriteFile(path, []byte(content), 0o644)
}

func patchServerMain(path string, m Model) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    s := string(b)

    importMarker := "// scaffold:imports (DO NOT REMOVE)"
    routeMarker := "// scaffold:routes (DO NOT REMOVE)"

    importLine := fmt.Sprintf("\t%q\n", fmt.Sprintf("%s/gen/%s/v1/%sv1connect", m.Module, toSnake(m.NameLower), toSnake(m.NameLower)))
    if strings.Contains(s, importLine) || strings.Contains(s, fmt.Sprintf(`"%s/gen/%s/v1/%sv1connect"`, m.Module, toSnake(m.NameLower), toSnake(m.NameLower))) {
        // already imported
    } else {
        if !strings.Contains(s, importMarker) {
            return fmt.Errorf("missing import marker in %s: %s", path, importMarker)
        }
        s = strings.Replace(s, importMarker, importMarker+"\n"+importLine, 1)
    }

    routeSnippet := fmt.Sprintf(`
    // %s scaffold
    %sRepo := memory.New%[1]sRepository()
    %sUC := usecase.New%[1]sUsecase(%sRepo)
    %sHandler := grpcadapter.New%[1]sHandler(%sUC)
    %sPath, %sH := %sv1connect.New%[1]sServiceHandler(%sHandler)
    mux.Handle(%sPath, %sH)
`, m.Name, toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower), toSnake(m.NameLower))

    if strings.Contains(s, fmt.Sprintf("New%sServiceHandler", m.Name)) {
        // already registered
    } else {
        if !strings.Contains(s, routeMarker) {
            return fmt.Errorf("missing route marker in %s: %s", path, routeMarker)
        }
        s = strings.Replace(s, routeMarker, routeMarker+"\n"+routeSnippet, 1)
    }

    return os.WriteFile(path, []byte(s), 0o644)
}

func isPascal(s string) bool {
    return regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`).MatchString(s)
}

func isLowerIdent(s string) bool {
    return regexp.MustCompile(`^[a-z][a-zA-Z0-9_]*$`).MatchString(s)
}

func toSnake(s string) string {
    var out []rune
    for i, r := range s {
        if i > 0 && r >= 'A' && r <= 'Z' {
            out = append(out, '_')
        }
        out = append(out, rune(strings.ToLower(string(r))[0]))
    }
    return string(out)
}

func pluralizeSnake(s string) string {
    if strings.HasSuffix(s, "s") {
        return s
    }
    return s + "s"
}

func toPascal(s string) string {
    parts := strings.Split(toSnake(s), "_")
    var b strings.Builder
    for _, p := range parts {
        if p == "" {
            continue
        }
        b.WriteString(strings.ToUpper(p[:1]))
        if len(p) > 1 {
            b.WriteString(p[1:])
        }
    }
    return b.String()
}

const protoTmpl = `syntax = "proto3";

package {{.ProtoPackage}};

option go_package = "{{.GoPackagePath}}";

service {{.Name}}Service {
  rpc Create{{.Name}}(Create{{.Name}}Request) returns (Create{{.Name}}Response) {}
  rpc Get{{.Name}}(Get{{.Name}}Request) returns (Get{{.Name}}Response) {}
  rpc List{{.Name}}s(List{{.Name}}sRequest) returns (List{{.Name}}sResponse) {}
  rpc Update{{.Name}}(Update{{.Name}}Request) returns (Update{{.Name}}Response) {}
  rpc Delete{{.Name}}(Delete{{.Name}}Request) returns (Delete{{.Name}}Response) {}
}

message {{.Name}} {
  int64 id = 1;
{{- range $i, $f := .Fields }}
  {{$f.ProtoType}} {{$f.JSONName}} = {{add2 $i}};
{{- end }}
}
message Create{{.Name}}Request {
{{- range $i, $f := .Fields }}
  {{$f.ProtoType}} {{$f.JSONName}} = {{inc $i}};
{{- end }}
}
message Create{{.Name}}Response { {{.Name}} {{.NameLower}} = 1; }

message Get{{.Name}}Request { int64 id = 1; }
message Get{{.Name}}Response { {{.Name}} {{.NameLower}} = 1; }

message List{{.Name}}sRequest {}
message List{{.Name}}sResponse { repeated {{.Name}} {{.NameLower}}s = 1; }

message Update{{.Name}}Request {
  int64 id = 1;
{{- range $i, $f := .Fields }}
  {{$f.ProtoType}} {{$f.JSONName}} = {{add2 $i}};
{{- end }}
}
message Update{{.Name}}Response { {{.Name}} {{.NameLower}} = 1; }

message Delete{{.Name}}Request { int64 id = 1; }
message Delete{{.Name}}Response {}
`

const usecaseTmpl = `package usecase

import "context"

type {{.Name}} struct {
    ID int64
{{- range .Fields }}
    {{.GoName}} {{if eq .ProtoType "int32"}}int32{{else if eq .ProtoType "int64"}}int64{{else if eq .ProtoType "bool"}}bool{{else}}string{{end}}
{{- end }}
}

type {{.Name}}Repository interface {
    Create(ctx context.Context, in *{{.Name}}) (*{{.Name}}, error)
    Get(ctx context.Context, id int64) (*{{.Name}}, error)
    List(ctx context.Context) ([]*{{.Name}}, error)
    Update(ctx context.Context, in *{{.Name}}) (*{{.Name}}, error)
    Delete(ctx context.Context, id int64) error
}

type {{.Name}}Usecase struct {
    repo {{.Name}}Repository
}

func New{{.Name}}Usecase(repo {{.Name}}Repository) *{{.Name}}Usecase {
    return &{{.Name}}Usecase{repo: repo}
}

func (u *{{.Name}}Usecase) Create(ctx context.Context, in *{{.Name}}) (*{{.Name}}, error) {
    return u.repo.Create(ctx, in)
}
func (u *{{.Name}}Usecase) Get(ctx context.Context, id int64) (*{{.Name}}, error) {
    return u.repo.Get(ctx, id)
}
func (u *{{.Name}}Usecase) List(ctx context.Context) ([]*{{.Name}}, error) {
    return u.repo.List(ctx)
}
func (u *{{.Name}}Usecase) Update(ctx context.Context, in *{{.Name}}) (*{{.Name}}, error) {
    return u.repo.Update(ctx, in)
}
func (u *{{.Name}}Usecase) Delete(ctx context.Context, id int64) error {
    return u.repo.Delete(ctx, id)
}
`

const handlerTmpl = `package grpc

import (
    "context"

    "connectrpc.com/connect"
    {{.GoPkgName}} "{{.Module}}/gen/{{.NameLower}}/v1"
    "{{.Module}}/internal/usecase"
)

type {{.Name}}Handler struct {
    uc *usecase.{{.Name}}Usecase
}

func New{{.Name}}Handler(uc *usecase.{{.Name}}Usecase) *{{.Name}}Handler {
    return &{{.Name}}Handler{uc: uc}
}

func (h *{{.Name}}Handler) Create{{.Name}}(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.Create{{.Name}}Request],
) (*connect.Response[{{.GoPkgName}}.Create{{.Name}}Response], error) {
    in := &usecase.{{.Name}}{
{{- range .Fields }}
        {{.GoName}}: req.Msg.Get{{.GoName}}(),
{{- end }}
    }
    out, err := h.uc.Create(ctx, in)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    res := connect.NewResponse(&{{.GoPkgName}}.Create{{.Name}}Response{
        {{.NameLower}}: toProto{{.Name}}(out),
    })
    return res, nil
}

func (h *{{.Name}}Handler) Get{{.Name}}(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.Get{{.Name}}Request],
) (*connect.Response[{{.GoPkgName}}.Get{{.Name}}Response], error) {
    out, err := h.uc.Get(ctx, req.Msg.GetId())
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    return connect.NewResponse(&{{.GoPkgName}}.Get{{.Name}}Response{ {{.NameLower}}: toProto{{.Name}}(out) }), nil
}

func (h *{{.Name}}Handler) List{{.Name}}s(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.List{{.Name}}sRequest],
) (*connect.Response[{{.GoPkgName}}.List{{.Name}}sResponse], error) {
    items, err := h.uc.List(ctx)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    out := make([]*{{.GoPkgName}}.{{.Name}}, 0, len(items))
    for _, it := range items {
        out = append(out, toProto{{.Name}}(it))
    }
    return connect.NewResponse(&{{.GoPkgName}}.List{{.Name}}sResponse{ {{.NameLower}}s: out }), nil
}

func (h *{{.Name}}Handler) Update{{.Name}}(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.Update{{.Name}}Request],
) (*connect.Response[{{.GoPkgName}}.Update{{.Name}}Response], error) {
    in := &usecase.{{.Name}}{
        ID: req.Msg.GetId(),
{{- range .Fields }}
        {{.GoName}}: req.Msg.Get{{.GoName}}(),
{{- end }}
    }
    out, err := h.uc.Update(ctx, in)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    return connect.NewResponse(&{{.GoPkgName}}.Update{{.Name}}Response{ {{.NameLower}}: toProto{{.Name}}(out) }), nil
}

func (h *{{.Name}}Handler) Delete{{.Name}}(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.Delete{{.Name}}Request],
) (*connect.Response[{{.GoPkgName}}.Delete{{.Name}}Response], error) {
    if err := h.uc.Delete(ctx, req.Msg.GetId()); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    return connect.NewResponse(&{{.GoPkgName}}.Delete{{.Name}}Response{}), nil
}

func toProto{{.Name}}(in *usecase.{{.Name}}) *{{.GoPkgName}}.{{.Name}} {
    if in == nil {
        return nil
    }
    return &{{.GoPkgName}}.{{.Name}}{
        Id: in.ID,
{{- range .Fields }}
        {{.GoName}}: in.{{.GoName}},
{{- end }}
    }
}
`

const repoMemoryTmpl = `package memory

import (
    "context"
    "sync"

    "{{.Module}}/internal/usecase"
)

type {{.Name}}Repository struct {
    mu   sync.Mutex
    seq  int64
    data map[int64]*usecase.{{.Name}}
}

func New{{.Name}}Repository() *{{.Name}}Repository {
    return &{{.Name}}Repository{data: map[int64]*usecase.{{.Name}}{}}
}

func (r *{{.Name}}Repository) Create(ctx context.Context, in *usecase.{{.Name}}) (*usecase.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.seq++
    cp := *in
    cp.ID = r.seq
    r.data[cp.ID] = &cp
    return &cp, nil
}

func (r *{{.Name}}Repository) Get(ctx context.Context, id int64) (*usecase.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    v, ok := r.data[id]
    if !ok {
        return nil, nil
    }
    cp := *v
    return &cp, nil
}

func (r *{{.Name}}Repository) List(ctx context.Context) ([]*usecase.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    out := make([]*usecase.{{.Name}}, 0, len(r.data))
    for _, v := range r.data {
        cp := *v
        out = append(out, &cp)
    }
    return out, nil
}

func (r *{{.Name}}Repository) Update(ctx context.Context, in *usecase.{{.Name}}) (*usecase.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.data[in.ID]; !ok {
        return nil, nil
    }
    cp := *in
    r.data[cp.ID] = &cp
    return &cp, nil
}

func (r *{{.Name}}Repository) Delete(ctx context.Context, id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.data, id)
    return nil
}
`

const repoMySQLTmpl = `package mysql

import (
    "context"
    "errors"

    "{{.Module}}/internal/usecase"
)

type {{.Name}}Repository struct{}

func New{{.Name}}Repository() *{{.Name}}Repository { return &{{.Name}}Repository{} }

var errNotImplemented = errors.New("mysql repository not implemented yet")

func (r *{{.Name}}Repository) Create(ctx context.Context, in *usecase.{{.Name}}) (*usecase.{{.Name}}, error) {
    return nil, errNotImplemented
}
func (r *{{.Name}}Repository) Get(ctx context.Context, id int64) (*usecase.{{.Name}}, error) {
    return nil, errNotImplemented
}
func (r *{{.Name}}Repository) List(ctx context.Context) ([]*usecase.{{.Name}}, error) {
    return nil, errNotImplemented
}
func (r *{{.Name}}Repository) Update(ctx context.Context, in *usecase.{{.Name}}) (*usecase.{{.Name}}, error) {
    return nil, errNotImplemented
}
func (r *{{.Name}}Repository) Delete(ctx context.Context, id int64) error {
    return errNotImplemented
}
`

const schemaTmpl = `-- {{.Name}} table
CREATE TABLE {{.Table}} (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
{{- range .Fields }}
  {{.JSONName}} {{.SQLType}} NOT NULL,
{{- end }}
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
`
