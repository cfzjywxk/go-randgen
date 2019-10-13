package view

import (
	"encoding/json"
	"github.com/pingcap/go-randgen/grammar"
	"github.com/pingcap/go-randgen/grammar/yacc_parser"
	"net/http"
)

type ProductionView struct {
	Number int  `json:"number"`
	Head   string  `json:"head"`
	Alter  []*SeqView  `json:"alter"`
}

type SeqView struct {
	Content string `json:"content"`
	Fanout  []int  `json:"fanout"`
}

type arraySet struct {
	arr  []int
	set  map[int]bool
}

func newArraySet() *arraySet {
	return &arraySet{arr: make([]int, 0), set: make(map[int]bool)}
}

func (a *arraySet) add(num int)  {
	if !a.set[num] {
		a.arr = append(a.arr, num)
		a.set[num] = true
	}
}

func productionToJson(productions []*yacc_parser.Production,
	pMap map[string]*yacc_parser.Production) ([]byte, error) {
	pViews := make([]*ProductionView, 0)

	for _, production := range productions {
		seqs := production.Alter
		seqViews := make([]*SeqView, 0, len(seqs))
		seqSet := make(map[string]bool)
		for _, seq := range seqs {
			content := seq.String()
			if _, ok := seqSet[content]; ok {
				// compress the seq set
				continue
			}
			seqSet[content] = true

			fanout := newArraySet()
			for _, item := range seq.Items {
				if yacc_parser.NonTerminalInMap(pMap, item) {
					fanout.add(pMap[item.ToString()].Number)
				}
			}

			seqViews = append(seqViews, &SeqView{Content: content, Fanout: fanout.arr})
		}

		pViews = append(pViews, &ProductionView{
			Number: production.Number,
			Head:   production.Head.ToString(),
			Alter:  seqViews,
		})
	}

	jsonBytes, err := json.Marshal(pViews)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

func Graph(yy string) (http.HandlerFunc, error) {
	_, productions, pMap, err := grammar.Parse(yy)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := productionToJson(productions, pMap)
	if err != nil {
		return nil, err
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(jsonBytes)
	}, nil
}
