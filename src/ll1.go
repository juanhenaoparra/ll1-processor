package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Grammar struct {
	Order          []string            `json:"order"`
	ProductionsSet map[string][]string `json:"productions_set"`
}

type LL1 struct {
	First  map[string][]string `json:"first"`
	Follow map[string][]string `json:"follow"`
}

type LL1Response struct {
	Grammar *Grammar `json:"grammar,omitempty"`
	Result  *LL1     `json:"result,omitempty"`
}

// LL1Process process the ll1 endpoint request
func LL1Process(w http.ResponseWriter, r *http.Request) {
	req := &Grammar{}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, "corrupted body payload"))
		return
	}

	defer CloseOrLog(r.Body)

	err = json.Unmarshal(bodyBytes, req)
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, "invalid body"))
		return
	}

	err = req.RemoveLeftRecursion()
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, fmt.Sprintf("remove left recursion failed: %s", err.Error())))
		return
	}

	ll1Response, err := req.ValidateLL1()
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, fmt.Sprintf("validate ll1 failed: %s", err.Error())))
		return
	}

	b, err := json.Marshal(ll1Response)
	if err != nil {
		rend(w, r, NewAPIError(http.StatusInternalServerError, "corrupted ll1 response body"))
		return
	}

	w.Write(b)
}

const (
	LambdaSymbol string = "λ"
)

var (
	ErrProductionSetAlreadyExists = errors.New("production set already exists")
	ErrProductionsSetNotFound     = errors.New("productions set not found")
	ErrProductionIndexNotFound    = errors.New("production index not found")
)

func (g *Grammar) GetIndexOfNonTerminal(nonterminal string) int {
	for i, nonterm := range g.Order {
		if nonterm == nonterminal {
			return i
		}
	}

	return -1
}

func (g *Grammar) GetIndexOfProduction(nonterminal, production string) int {
	for i, prod := range g.ProductionsSet[production] {
		if prod == production {
			return i
		}
	}

	return -1
}

func (g *Grammar) AddProductionGroup(nonterminal string, productions []string) {
	nonterminal = strings.TrimSpace(nonterminal)
	index := g.GetIndexOfNonTerminal(nonterminal)

	foundProductions, ok := g.ProductionsSet[nonterminal]
	if ok || index != -1 {
		g.ProductionsSet[nonterminal] = UnionStringSet(foundProductions, productions)
		return
	}

	g.Order = append(g.Order, nonterminal)
	g.ProductionsSet[nonterminal] = productions
}

func (g *Grammar) HasLeftRecursion(prefix string, productions []string) bool {
	for _, production := range productions {
		if strings.HasPrefix(production, prefix) {
			return true
		}
	}

	return false
}

func (g *Grammar) RemoveLeftRecursion() error {
	for nonterminal, productions := range g.ProductionsSet {
		if !g.HasLeftRecursion(nonterminal, productions) {
			continue
		}

		betaProductions := make([]string, 0, len(productions))

		for _, production := range productions {
			nonterminalPrim := " " + nonterminal + "'"

			if production == LambdaSymbol {
				continue
			}

			if !strings.HasPrefix(production, nonterminal) {
				betaProductions = append(betaProductions, strings.TrimSpace(production+nonterminalPrim))
				continue
			}

			newProduction := strings.TrimSpace(strings.TrimPrefix(production, nonterminal) + nonterminalPrim)
			g.AddProductionGroup(nonterminalPrim, []string{newProduction, LambdaSymbol})
		}

		g.ProductionsSet[nonterminal] = betaProductions
	}

	return nil
}

func (g *Grammar) ValidateLL1() (*LL1Response, error) {
	ll1Response := &LL1Response{
		Grammar: g,
		Result:  &LL1{},
	}

	first, err := g.ComputeFirst()
	if err != nil {
		return nil, err
	}

	ll1Response.Result.First = first

	return ll1Response, nil
}

func IsTerminal(set map[string][]string, v string) bool {
	_, ok := set[v]

	return !ok
}

func (g *Grammar) ComputeFirst() (map[string][]string, error) {
	first := make(map[string][]string)

	for _, nonterminal := range g.Order {
		nonterminalFirst, err := GetFirstOfNonterminal(g.ProductionsSet, nonterminal)
		if err != nil {
			return nil, err
		}

		first[nonterminal] = nonterminalFirst
	}

	return first, nil
}

func GetFirstOfNonterminal(set map[string][]string, nonterminal string) ([]string, error) {
	productions, ok := set[nonterminal]
	if !ok {
		return []string{}, ErrProductionsSetNotFound
	}

	firstSet := make([]string, 0, len(productions))

	for _, production := range productions {
		if production == LambdaSymbol {
			firstSet = append(firstSet, LambdaSymbol)
			continue
		}

		allWords := strings.Split(production, " ")

		for _, word := range allWords {
			if IsTerminal(set, word) {
				firstSet = append(firstSet, word)
				break
			}

			foundFirstSet, err := GetFirstOfNonterminal(set, word)
			if err != nil {
				return firstSet, err
			}

			firstSet = append(firstSet, foundFirstSet...)
			break
		}
	}

	firstSet = UnionStringSet(firstSet, firstSet)

	return firstSet, nil
}

func ContainsAny(l []string, v string) (int, bool) {
	for i, item := range l {
		if item == v {
			return i, true
		}
	}

	return -1, false
}

func UnionStringSet(set []string, valuesToAdd []string) []string {
	setMap := make(map[string]bool)

	for _, v := range set {
		setMap[v] = true
	}

	for _, v := range valuesToAdd {
		if _, ok := setMap[v]; !ok {
			set = append(set, v)
			setMap[v] = true
		}
	}

	return set
}
