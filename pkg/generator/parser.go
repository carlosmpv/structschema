package generator

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/ettle/strcase"
)

type IDRef struct {
	Value json.RawMessage
}

func (idr *IDRef) GetIDs() ([]NodeID, error) {
	var ids []NodeID

	// Tenta deserializar como string única
	var singleID struct {
		ID NodeID `json:"@id"`
	}
	if err := json.Unmarshal(idr.Value, &singleID); err == nil {
		ids = append(ids, singleID.ID)
		return ids, nil
	}

	// Tenta deserializar como array de objetos
	var objects []struct {
		ID NodeID `json:"@id"`
	}
	if err := json.Unmarshal(idr.Value, &objects); err == nil {
		for _, obj := range objects {
			ids = append(ids, obj.ID)
		}
		return ids, nil
	}

	// Tenta deserializar como array de strings
	var stringArray []NodeID
	if err := json.Unmarshal(idr.Value, &stringArray); err == nil {
		return stringArray, nil
	}

	return nil, fmt.Errorf("invalid IDRef format")
}

func (idr *IDRef) UnmarshalJSON(data []byte) error {
	idr.Value = make(json.RawMessage, len(data))
	copy(idr.Value, data)
	return nil
}

func (idr IDRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(idr.Value)
}

// StringOrArray lida com campos que podem ser string ou []string
type StringOrArray struct {
	Values []string
}

// UnmarshalJSON para StringOrArray
func (sa *StringOrArray) UnmarshalJSON(data []byte) error {
	// Tenta deserializar como string simples
	var singleString string
	if err := json.Unmarshal(data, &singleString); err == nil {
		sa.Values = []string{singleString}
		return nil
	}

	// Tenta deserializar como array de strings
	var stringArray []string
	if err := json.Unmarshal(data, &stringArray); err == nil {
		sa.Values = stringArray
		return nil
	}

	return fmt.Errorf("cannot unmarshal StringOrArray: %s", string(data))
}

// MarshalJSON para StringOrArray
func (sa StringOrArray) MarshalJSON() ([]byte, error) {
	if len(sa.Values) == 1 {
		return json.Marshal(sa.Values[0])
	}
	return json.Marshal(sa.Values)
}

// Métodos auxiliares para facilitar o uso
func (sa StringOrArray) First() string {
	if len(sa.Values) > 0 {
		return sa.Values[0]
	}
	return ""
}

func (sa StringOrArray) Contains(value string) bool {
	for _, v := range sa.Values {
		if v == value {
			return true
		}
	}
	return false
}

type LanguageString struct {
	Value    string `json:"@value,omitempty"`
	Language string `json:"@language,omitempty"`
}

// UnmarshalJSON para LanguageString
func (ls *LanguageString) UnmarshalJSON(data []byte) error {
	// Tenta deserializar como string simples
	var simpleString string
	if err := json.Unmarshal(data, &simpleString); err == nil {
		ls.Value = simpleString
		return nil
	}

	// Tenta deserializar como objeto com @value e @language
	var obj struct {
		Value    string `json:"@value"`
		Language string `json:"@language,omitempty"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		ls.Value = obj.Value
		ls.Language = obj.Language
		return nil
	}

	return fmt.Errorf("cannot unmarshal LanguageString: %s", string(data))
}

// MarshalJSON para LanguageString
func (ls LanguageString) MarshalJSON() ([]byte, error) {
	// Se não tem language, serializa como string simples
	if ls.Language == "" {
		return json.Marshal(ls.Value)
	}

	// Se tem language, serializa como objeto
	return json.Marshal(struct {
		Value    string `json:"@value"`
		Language string `json:"@language,omitempty"`
	}{
		Value:    ls.Value,
		Language: ls.Language,
	})
}

// String retorna apenas o valor (ignorando o idioma)
func (ls LanguageString) String() string {
	return ls.Value
}

type SchemaContext struct {
	Bibo           string `json:"bibo"`
	Brick          string `json:"brick"`
	CmnsCls        string `json:"cmns-cls"`
	CmnsCol        string `json:"cmns-col"`
	CmnsDt         string `json:"cmns-dt"`
	CmnsGe         string `json:"cmns-ge"`
	CmnsID         string `json:"cmns-id"`
	CmnsLoc        string `json:"cmns-loc"`
	CmnsQ          string `json:"cmns-q"`
	CmnsTxt        string `json:"cmns-txt"`
	Csvw           string `json:"csvw"`
	Dc             string `json:"dc"`
	Dcam           string `json:"dcam"`
	Dcat           string `json:"dcat"`
	Dcmitype       string `json:"dcmitype"`
	Dct            string `json:"dct"`
	Dcterms        string `json:"dcterms"`
	Dctype         string `json:"dctype"`
	Doap           string `json:"doap"`
	Eli            string `json:"eli"`
	FiboBeCorpCorp string `json:"fibo-be-corp-corp"`
	FiboBeGeGe     string `json:"fibo-be-ge-ge"`
	FiboBeLeCb     string `json:"fibo-be-le-cb"`
	FiboBeLeLp     string `json:"fibo-be-le-lp"`
	FiboBeNfpNfp   string `json:"fibo-be-nfp-nfp"`
	FiboBeOacCctl  string `json:"fibo-be-oac-cctl"`
	FiboFbcDaeDbt  string `json:"fibo-fbc-dae-dbt"`
	FiboFbcPasFpas string `json:"fibo-fbc-pas-fpas"`
	FiboFndAccCur  string `json:"fibo-fnd-acc-cur"`
	FiboFndAgrCtr  string `json:"fibo-fnd-agr-ctr"`
	FiboFndArrDoc  string `json:"fibo-fnd-arr-doc"`
	FiboFndArrLif  string `json:"fibo-fnd-arr-lif"`
	FiboFndDtOc    string `json:"fibo-fnd-dt-oc"`
	FiboFndOrgOrg  string `json:"fibo-fnd-org-org"`
	FiboFndPasPas  string `json:"fibo-fnd-pas-pas"`
	FiboFndPlcAdr  string `json:"fibo-fnd-plc-adr"`
	FiboFndPlcFac  string `json:"fibo-fnd-plc-fac"`
	FiboFndPlcLoc  string `json:"fibo-fnd-plc-loc"`
	FiboFndPtyPty  string `json:"fibo-fnd-pty-pty"`
	FiboFndRelRel  string `json:"fibo-fnd-rel-rel"`
	FiboPayPsPs    string `json:"fibo-pay-ps-ps"`
	Foaf           string `json:"foaf"`
	GleifL1        string `json:"gleif-L1"`
	Gs1            string `json:"gs1"`
	Hydra          string `json:"hydra"`
	Lcc31661       string `json:"lcc-3166-1"`
	Lcc4217        string `json:"lcc-4217"`
	LccCr          string `json:"lcc-cr"`
	LccLr          string `json:"lcc-lr"`
	Lrmoo          string `json:"lrmoo"`
	Mo             string `json:"mo"`
	Odrl           string `json:"odrl"`
	Org            string `json:"org"`
	Owl            string `json:"owl"`
	Prof           string `json:"prof"`
	Prov           string `json:"prov"`
	Qb             string `json:"qb"`
	Rdf            string `json:"rdf"`
	Rdfs           string `json:"rdfs"`
	Sarif          string `json:"sarif"`
	Schema         string `json:"schema"`
	Sh             string `json:"sh"`
	Skos           string `json:"skos"`
	Snomed         string `json:"snomed"`
	Sosa           string `json:"sosa"`
	Ssn            string `json:"ssn"`
	Time           string `json:"time"`
	Unece          string `json:"unece"`
	Vann           string `json:"vann"`
	Vcard          string `json:"vcard"`
	Void           string `json:"void"`
	Xsd            string `json:"xsd"`
}

type NodeID string

func isLower(s string) bool {
	return strings.ToLower(s) == s
}

func (nid NodeID) ToKebab() string {
	idParts := strings.Split(path.Base(string(nid)), ":")
	typeNameParts := make([]string, len(idParts))
	for i, keyword := range idParts {
		if isLower(keyword[:1]) && keyword != "schema" {
			typeNameParts[i] = "Prop" + strcase.ToPascal(keyword)
		} else {
			typeNameParts[i] = strcase.ToPascal(keyword)
		}
	}
	return strings.Join(typeNameParts, "")
}

type SchemaGraphNode struct {
	ID          NodeID         `json:"@id"`
	Type        StringOrArray  `json:"@type"`
	RdfsComment LanguageString `json:"rdfs:comment,omitempty"` // will generates comments on structs
	RdfsLabel   LanguageString `json:"rdfs:label,omitempty"`

	// To be filtered
	SchemaInverseOf    *IDRef `json:"schema:inverseOf,omitempty"`    // May be used to remove inversed definitions
	SchemaSupersededBy *IDRef `json:"schema:supersededBy,omitempty"` // May be used to remove deprecated definitions

	// Child node pointers
	RdfsSubPropertyOf    *IDRef `json:"rdfs:subPropertyOf,omitempty"`
	SchemaDomainIncludes *IDRef `json:"schema:domainIncludes,omitempty"`
	SchemaRangeIncludes  *IDRef `json:"schema:rangeIncludes,omitempty"`
	RdfsSubClassOf       *IDRef `json:"rdfs:subClassOf,omitempty"`

	// Childs maps @id's to:
	// - rdfs:subPropertyOf
	// - schema:domainIncludes
	// - schema:rangeIncludes
	// its a pointer for it to be copied by subClasses
	so     *SchemaOrg
	Childs []NodeID // Populate by search RdfsSubPropertyOf

	// The atributes below are feasible but unecessary for structs generation
	// OwlEquivalentProperty *IDRef `json:"owl:equivalentProperty,omitempty"`
	// SchemaContributor     *IDRef `json:"schema:contributor,omitempty"`
	// SchemaSource          *IDRef `json:"schema:source,omitempty"`
	// SkosExactMatch        *IDRef `json:"skos:exactMatch,omitempty"`
	// SkosCloseMatch        *IDRef `json:"skos:closeMatch,omitempty"`
	// OwlEquivalentClass    *IDRef `json:"owl:equivalentClass,omitempty"`
	// SchemaSameAs          *IDRef `json:"schema:sameAs,omitempty"`
	// RdfsSeeAlso           *IDRef `json:"rdfs:seeAlso,omitempty"`
	// OwlDisjointWith       *IDRef `json:"owl:disjointWith,omitempty"`
}

type SchemaOrg struct {
	ExpandedCounter int
	Context         SchemaContext      `json:"@context"`
	Graph           []*SchemaGraphNode `json:"@graph"`

	nodeMap map[NodeID]*SchemaGraphNode
}

func (so *SchemaOrg) FindDef(sID NodeID) *SchemaGraphNode {
	for _, def := range so.Graph {
		if def.ID == sID {
			def.so = so
			return def
		}
	}

	return nil
}

func (so *SchemaOrg) GenFile(ident string) *ast.File {
	decls := []ast.Decl{}

	for _, node := range so.nodeMap {
		decls = append(decls, node.GenDecl())
	}

	return &ast.File{
		Name:  ast.NewIdent(ident),
		Decls: decls,
	}
}

func (node *SchemaGraphNode) GenDecl() *ast.GenDecl {
	typeName := node.ID.ToKebab()

	fieldList := []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent("Context")},
			Type:  ast.NewIdent("string"),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"@context\"`",
			},
		},
		{
			Names: []*ast.Ident{ast.NewIdent("Type")},
			Type:  ast.NewIdent("string"),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"@type\"`",
			},
		},
	}

	for _, child := range node.Childs {
		kebabID := child.ToKebab()
		fieldList = append(fieldList, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(kebabID)},
			Type:  ast.NewIdent("[]*" + kebabID),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"" + string(child) + "\"`",
			},
		})
	}

	return &ast.GenDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Slash: token.NoPos,
					Text:  fmt.Sprintf("// %s", strings.ReplaceAll(node.RdfsComment.String(), "\n", "\n//")),
				},
			},
		},
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(typeName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fieldList,
					},
				},
			},
		},
	}
}

func Parse(r io.Reader) (*SchemaOrg, error) {
	var data SchemaOrg
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	data.nodeMap = map[NodeID]*SchemaGraphNode{}
	for _, node := range data.Graph {
		data.nodeMap[node.ID] = node
	}

	for _, node := range data.Graph {
		if node.RdfsSubPropertyOf != nil {
			ids, err := node.RdfsSubPropertyOf.GetIDs()
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				if _, ok := data.nodeMap[id]; !ok {
					continue
				}

				if data.nodeMap[id].Childs == nil {
					data.nodeMap[id].Childs = []NodeID{}
				}

				if !slices.Contains(data.nodeMap[id].Childs, node.ID) {
					data.nodeMap[id].Childs = append(data.nodeMap[id].Childs, node.ID)
				}
			}
		}
		if node.SchemaDomainIncludes != nil {
			ids, err := node.SchemaDomainIncludes.GetIDs()
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				if _, ok := data.nodeMap[id]; !ok {
					continue
				}

				if data.nodeMap[id].Childs == nil {
					data.nodeMap[id].Childs = []NodeID{}
				}

				if !slices.Contains(data.nodeMap[id].Childs, node.ID) {
					data.nodeMap[id].Childs = append(data.nodeMap[id].Childs, node.ID)
				}
			}
		}
		if node.SchemaRangeIncludes != nil {
			ids, err := node.SchemaRangeIncludes.GetIDs()
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				if node.Childs == nil {
					node.Childs = []NodeID{}
				}

				if !slices.Contains(data.nodeMap[id].Childs, id) {
					node.Childs = append(node.Childs, id)
				}
			}
		}
	}

	for _, node := range data.Graph {
		if node.RdfsSubClassOf != nil {
			ids, err := node.RdfsSubClassOf.GetIDs()
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				if node.Childs == nil {
					node.Childs = []NodeID{}
				}

				for _, child := range data.nodeMap[id].Childs {
					if !slices.Contains(node.Childs, child) {
						node.Childs = append(node.Childs, child)
					}
				}
			}
		}
	}

	return &data, nil
}
