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
	DBName    string
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
	var withMemory bool
	flag.StringVar(&name, "name", "", "Entity name in PascalCase, e.g. User")
	flag.StringVar(&fields, "fields", "", `Fields, e.g. "name:string email:string age:int"`)
	flag.BoolVar(&withMemory, "with-memory", false, "also generate in-memory repository implementation")
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
	// domain entity
	if err := writeFromTemplate("domain", filepath.Join("internal", "domain", "entity", m.NameLower+".go"), domainTmpl, m); err != nil {
		exitErr(err)
	}
	if err := writeFromTemplate("usecase", filepath.Join("internal", "usecase", m.NameLower+"_usecase.go"), usecaseTmpl, m); err != nil {
		exitErr(err)
	}
	if err := writeFromTemplate("handler", filepath.Join("internal", "adapter", "grpc", m.NameLower+"_handler.go"), handlerTmpl, m); err != nil {
		exitErr(err)
	}
	if err := writeFromTemplate("domain-repo", filepath.Join("internal", "domain", "repository", m.NameLower+"_repository.go"), domainRepoTmpl, m); err != nil {
		exitErr(err)
	}
	if withMemory {
		if err := writeFromTemplate("repo-mem", filepath.Join("internal", "adapter", "repository", "memory", m.NameLower+"_repository.go"), repoMemoryTmpl, m); err != nil {
			exitErr(err)
		}
	}
	if err := writeFromTemplate("repo-mysql", filepath.Join("internal", "adapter", "repository", "mysql", m.NameLower+"_repository.go"), repoMySQLTmpl, m); err != nil {
		exitErr(err)
	}

	if err := ensureSchemaSQL(filepath.Join("db", "schema.sql"), schemaTmpl, m); err != nil {
		exitErr(err)
	}
	// add per-entity route registrar (registry-based; main.go stays unchanged)
	if err := writeFromTemplate("routes", filepath.Join("internal", "adapter", "grpc", m.NameLower+"_routes.go"), routesTmpl, m); err != nil {
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
		dbname := toSnake(name)
		// Disallow reserved SQL identifiers to avoid invalid DDL
		switch dbname {
		case "text", "order", "group", "value":
			return nil, fmt.Errorf("field %s: %q is a reserved SQL identifier; choose a different name (e.g., %s_col)", name, dbname, dbname)
		}
		out = append(out, Field{
			Name:      name,
			GoName:    toPascal(name),
			ProtoType: protoType,
			SQLType:   sqlType,
			JSONName:  toSnake(name),
			DBName:    dbname,
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
	case "int8":
		return "int32", "TINYINT", nil
	case "int64":
		return "int64", "BIGINT", nil
	case "uint8":
		return "uint32", "TINYINT UNSIGNED", nil
	case "uint32":
		return "uint32", "INT UNSIGNED", nil
	case "uint64":
		return "uint64", "BIGINT UNSIGNED", nil
	case "bool":
		return "bool", "TINYINT(1)", nil
	default:
		return "", "", fmt.Errorf("unknown type %q (supported: string,text,int,int8,int32,int64,uint8,uint32,uint64,bool)", t)
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
		"inc":  func(i int) int { return i + 1 },
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

// legacy patchServerMain removed in favor of registry-based routes

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

const domainTmpl = `package entity

type {{.Name}} struct {
    ID int64
{{- range .Fields }}
    {{.GoName}} {{if eq .ProtoType "int32"}}int32{{else if eq .ProtoType "int64"}}int64{{else if eq .ProtoType "uint32"}}uint32{{else if eq .ProtoType "uint64"}}uint64{{else if eq .ProtoType "bool"}}bool{{else}}string{{end}}
{{- end }}
}
`

const usecaseTmpl = `package usecase

import (
    "context"
    "{{.Module}}/internal/domain"
    "{{.Module}}/internal/domain/entity"
    domainrepo "{{.Module}}/internal/domain/repository"
)

type {{.Name}}Usecase struct {
    repo domainrepo.{{.Name}}Repository
}

func New{{.Name}}Usecase(repo domainrepo.{{.Name}}Repository) *{{.Name}}Usecase {
    return &{{.Name}}Usecase{repo: repo}
}

func (u *{{.Name}}Usecase) Create(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
    return u.repo.Create(ctx, in)
}
func (u *{{.Name}}Usecase) Get(ctx context.Context, id int64) (*entity.{{.Name}}, error) {
    return u.repo.Get(ctx, id)
}
func (u *{{.Name}}Usecase) List(ctx context.Context, p domain.ListParams) ([]*entity.{{.Name}}, error) {
    return u.repo.List(ctx, p)
}
func (u *{{.Name}}Usecase) Update(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
    return u.repo.Update(ctx, in)
}
func (u *{{.Name}}Usecase) Delete(ctx context.Context, id int64) error {
    return u.repo.Delete(ctx, id)
}
`

const domainRepoTmpl = `package repository

import (
    "context"
    "{{.Module}}/internal/domain"
    "{{.Module}}/internal/domain/entity"
)

type {{.Name}}Repository interface {
    Create(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error)
    Get(ctx context.Context, id int64) (*entity.{{.Name}}, error)
    List(ctx context.Context, p domain.ListParams) ([]*entity.{{.Name}}, error)
    Update(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error)
    Delete(ctx context.Context, id int64) error
}
`

const handlerTmpl = `package grpc

import (
    "context"

    "connectrpc.com/connect"
    {{.GoPkgName}} "{{.Module}}/gen/{{.NameLower}}/v1"
    "{{.Module}}/internal/domain"
    "{{.Module}}/internal/domain/entity"
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
    in := &entity.{{.Name}}{
{{- range .Fields }}
        {{.GoName}}: req.Msg.Get{{.GoName}}(),
{{- end }}
    }
    out, err := h.uc.Create(ctx, in)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    res := connect.NewResponse(&{{.GoPkgName}}.Create{{.Name}}Response{
        {{.Name}}: toProto{{.Name}}(out),
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
    return connect.NewResponse(&{{.GoPkgName}}.Get{{.Name}}Response{ {{.Name}}: toProto{{.Name}}(out) }), nil
}

func (h *{{.Name}}Handler) List{{.Name}}s(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.List{{.Name}}sRequest],
) (*connect.Response[{{.GoPkgName}}.List{{.Name}}sResponse], error) {
    items, err := h.uc.List(ctx, domain.ListParams{})
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    out := make([]*{{.GoPkgName}}.{{.Name}}, 0, len(items))
    for _, it := range items {
        out = append(out, toProto{{.Name}}(it))
    }
    return connect.NewResponse(&{{.GoPkgName}}.List{{.Name}}sResponse{ {{.Name}}s: out }), nil
}

func (h *{{.Name}}Handler) Update{{.Name}}(
    ctx context.Context,
    req *connect.Request[{{.GoPkgName}}.Update{{.Name}}Request],
) (*connect.Response[{{.GoPkgName}}.Update{{.Name}}Response], error) {
    in := &entity.{{.Name}}{
        ID: req.Msg.GetId(),
{{- range .Fields }}
        {{.GoName}}: req.Msg.Get{{.GoName}}(),
{{- end }}
    }
    out, err := h.uc.Update(ctx, in)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    return connect.NewResponse(&{{.GoPkgName}}.Update{{.Name}}Response{ {{.Name}}: toProto{{.Name}}(out) }), nil
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

func toProto{{.Name}}(in *entity.{{.Name}}) *{{.GoPkgName}}.{{.Name}} {
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

const routesTmpl = `package grpc

import (
    "net/http"

    {{.GoPkgName}}connect "{{.Module}}/gen/{{.NameLower}}/v1/{{.NameLower}}v1connect"
    mysqlrepo "{{.Module}}/internal/adapter/repository/mysql"
    "{{.Module}}/internal/usecase"
)

func init() { Add(register{{.Name}}) }

func register{{.Name}}(mux *http.ServeMux, deps Deps) {
    repo := mysqlrepo.New{{.Name}}Repository(deps.Gorm)
    uc := usecase.New{{.Name}}Usecase(repo)
    h := New{{.Name}}Handler(uc)
    path, handler := {{.GoPkgName}}connect.New{{.Name}}ServiceHandler(h)
    mux.Handle(path, handler)
}
`

const repoMemoryTmpl = `package memory

import (
    "context"
    "sync"

    "{{.Module}}/internal/domain"
    "{{.Module}}/internal/domain/entity"
    domainrepo "{{.Module}}/internal/domain/repository"
)

type {{.Name}}Repository struct {
    mu   sync.Mutex
    seq  int64
    data map[int64]*entity.{{.Name}}
}

func New{{.Name}}Repository() domainrepo.{{.Name}}Repository {
    return &{{.Name}}Repository{data: map[int64]*entity.{{.Name}}{}}
}

func (r *{{.Name}}Repository) Create(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.seq++
    cp := *in
    cp.ID = r.seq
    r.data[cp.ID] = &cp
    return &cp, nil
}

func (r *{{.Name}}Repository) Get(ctx context.Context, id int64) (*entity.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    v, ok := r.data[id]
    if !ok {
        return nil, nil
    }
    cp := *v
    return &cp, nil
}

func (r *{{.Name}}Repository) List(ctx context.Context, p domain.ListParams) ([]*entity.{{.Name}}, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    tmp := make([]*entity.{{.Name}}, 0, len(r.data))
    for _, v := range r.data { cp := *v; tmp = append(tmp, &cp) }
    p = p.Sanitize()
    start := p.Offset
    if start > len(tmp) { start = len(tmp) }
    end := start + p.Limit
    if end > len(tmp) { end = len(tmp) }
    out := make([]*entity.{{.Name}}, end-start)
    copy(out, tmp[start:end])
    return out, nil
}

func (r *{{.Name}}Repository) Update(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
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
    "time"

    "gorm.io/gorm"

    "{{.Module}}/internal/domain"
    "{{.Module}}/internal/domain/entity"
    domainrepo "{{.Module}}/internal/domain/repository"
)

type {{.Name}}Model struct {
    ID int64 ` + "`gorm:\"primaryKey;autoIncrement\"`" + `
{{- range .Fields }}
    {{.GoName}} {{if eq .ProtoType "int32"}}int32{{else if eq .ProtoType "int64"}}int64{{else if eq .ProtoType "uint32"}}uint32{{else if eq .ProtoType "uint64"}}uint64{{else if eq .ProtoType "bool"}}bool{{else}}string{{end}} ` + "`gorm:\"column:{{.DBName}};not null\"`" + `
{{- end }}
    CreatedAt time.Time ` + "`gorm:\"column:created_at;autoCreateTime\"`" + `
    UpdatedAt time.Time ` + "`gorm:\"column:updated_at;autoUpdateTime\"`" + `
}

func ({{.Name}}Model) TableName() string { return "{{.Table}}" }

type {{.Name}}Repository struct{ db *gorm.DB }

func New{{.Name}}Repository(db *gorm.DB) domainrepo.{{.Name}}Repository { return &{{.Name}}Repository{db: db} }

func (r *{{.Name}}Repository) Create(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
    m := {{.Name}}Model{
{{- range .Fields }}
        {{.GoName}}: in.{{.GoName}},
{{- end }}
    }
    if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
        return nil, err
    }
    out := *in
    out.ID = m.ID
    return &out, nil
}

func (r *{{.Name}}Repository) Get(ctx context.Context, id int64) (*entity.{{.Name}}, error) {
    var m {{.Name}}Model
    if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &entity.{{.Name}}{
        ID: m.ID,
{{- range .Fields }}
        {{.GoName}}: m.{{.GoName}},
{{- end }}
    }, nil
}

func (r *{{.Name}}Repository) List(ctx context.Context, p domain.ListParams) ([]*entity.{{.Name}}, error) {
    var rows []{{.Name}}Model
    p = p.Sanitize()
    q := r.db.WithContext(ctx).Order("id DESC").Offset(p.Offset).Limit(p.Limit)
    if err := q.Find(&rows).Error; err != nil {
        return nil, err
    }
    out := make([]*entity.{{.Name}}, 0, len(rows))
    for _, m := range rows {
        it := entity.{{.Name}}{
            ID: m.ID,
            {{- range .Fields }}
            {{.GoName}}: m.{{.GoName}},
            {{- end }}
        }
        out = append(out, &it)
    }
    return out, nil
}

func (r *{{.Name}}Repository) Update(ctx context.Context, in *entity.{{.Name}}) (*entity.{{.Name}}, error) {
    updates := map[string]interface{}{
{{- range .Fields }}
        "{{.DBName}}": in.{{.GoName}},
{{- end }}
        "updated_at": time.Now(),
    }
    if err := r.db.WithContext(ctx).Model(&{{.Name}}Model{}).Where("id = ?", in.ID).Updates(updates).Error; err != nil {
        return nil, err
    }
    return r.Get(ctx, in.ID)
}

func (r *{{.Name}}Repository) Delete(ctx context.Context, id int64) error {
    return r.db.WithContext(ctx).Delete(&{{.Name}}Model{}, id).Error
}
`

const schemaTmpl = `-- {{.Name}} table
CREATE TABLE {{.Table}} (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
{{- range .Fields }}
  {{.DBName}} {{.SQLType}} NOT NULL,
{{- end }}
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
`
