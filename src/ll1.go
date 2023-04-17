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
	First      map[string][]string `json:"first"`
	Follow     map[string][]string `json:"follow"`
	Prediction map[string][]string `json:"prediction"`
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
	DollarSymbol string = "$"
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

	follow, err := g.ComputeFollow(first)
	if err != nil {
		return nil, err
	}

	ll1Response.Result.First = first
	ll1Response.Result.Follow = follow
	ll1Response.Result.Prediction = g.ComputePredictionSet(first, follow)

	return ll1Response, nil
}

func (g *Grammar) ComputePredictionSet(first, follow map[string][]string) map[string][]string {
	predictionSet := map[string][]string{}

	for _, nonterminal := range g.Order {
		values := first[nonterminal]

		_, containsLambda := ContainsAny(values, LambdaSymbol)

		if containsLambda {
			values = follow[nonterminal]
		}

		predictionSet[nonterminal] = values
	}

	return predictionSet
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

		for i, word := range allWords {
			if IsTerminal(set, word) {
				firstSet = append(firstSet, word)
				break
			}

			foundFirstSet, err := GetFirstOfNonterminal(set, word)
			if err != nil {
				return firstSet, err
			}

			_, containsLambda := ContainsAny(foundFirstSet, LambdaSymbol)

			if containsLambda && i < len(allWords)-1 {
				firstSet = append(firstSet, RemoveElement(foundFirstSet, LambdaSymbol)...)
				continue
			}

			firstSet = append(firstSet, foundFirstSet...)
			break
		}
	}

	firstSet = UnionStringSet(firstSet, firstSet)

	return firstSet, nil
}

func (g *Grammar) ComputeFollow(firstSet map[string][]string) (map[string][]string, error) {
	follow := make(map[string][]string)

	for i, nonterminal := range g.Order {
		if i == 0 {
			follow[nonterminal] = []string{DollarSymbol}
		}

		foundFollow, err := GetFollowOfNonterminal(g.ProductionsSet, firstSet, follow, nonterminal)
		if err != nil {
			return nil, err
		}

		follow[nonterminal] = UnionStringSet(foundFollow, foundFollow)
	}

	for nonterminal, follows := range follow {
		follow[nonterminal] = UnionStringSet(follows, follows)
	}

	return follow, nil
}

func GetFollowingFromProduction(production string, value string) string {
	productionSplitted := strings.SplitN(production, value, 2)

	if len(production) <= 1 {
		return LambdaSymbol
	}

	return strings.Split(strings.TrimSpace(productionSplitted[1]), " ")[0]
}

func FindNonterminalOccurrences(set map[string][]string, nonterminalSearched string) map[string][]string {
	ocurrences := make(map[string][]string)

	for nonterminal, productions := range set {
		for _, production := range productions {
			if !ContainsWord(production, nonterminalSearched) {
				continue
			}

			if _, ok := ocurrences[nonterminal]; !ok {
				ocurrences[nonterminal] = []string{}
			}

			ocurrences[nonterminal] = append(ocurrences[nonterminal], GetFollowingFromProduction(production, nonterminalSearched))
		}
	}

	return ocurrences
}

func p(a any) string {
	b, _ := json.Marshal(a)
	return string(b)
}

func GetFollowOfNonterminal(set map[string][]string, first map[string][]string, currentFollows map[string][]string, nonterminal string) ([]string, error) {
	follows := currentFollows[nonterminal]
	foundNonterminalOcurrences := FindNonterminalOccurrences(set, nonterminal)

	fmt.Println("nt: ", nonterminal, ", occ: ", p(foundNonterminalOcurrences))

	for nt, productions := range foundNonterminalOcurrences {
		for _, production := range productions {
			isTerminal := IsTerminal(set, production)
			if isTerminal && production != "" {
				follows = append(follows, production)
				continue
			}

			if production == "" {
				follows = append(follows, currentFollows[nt]...)
				continue
			}

			if production == LambdaSymbol {
				recursiveFollows, err := GetFollowOfNonterminal(set, first, currentFollows, nt)
				if err != nil {
					return nil, err
				}

				follows = append(follows, recursiveFollows...)

				continue
			}

			firstOfProduction := first[production]

			follows = append(follows, RemoveElement(firstOfProduction, LambdaSymbol)...)

			if _, containsLambda := ContainsAny(firstOfProduction, LambdaSymbol); containsLambda {
				recursiveFollows, err := GetFollowOfNonterminal(set, first, currentFollows, nt)
				if err != nil {
					return nil, err
				}

				follows = append(follows, recursiveFollows...)

				continue
			}
		}
	}

	follows = UnionStringSet(follows, follows)
	currentFollows[nonterminal] = follows

	return follows, nil
}

func RemoveElement(l []string, v string) []string {
	set := make(map[string]bool)
	for _, value := range l {
		set[value] = true
	}

	delete(set, v)

	newList := make([]string, 0, len(set))
	for value := range set {
		newList = append(newList, value)
	}

	return newList
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
	unionSet := make([]string, 0, len(set)+len(valuesToAdd))

	for _, v := range set {
		if _, ok := setMap[v]; !ok {
			setMap[v] = true
			unionSet = append(unionSet, v)
		}
	}

	for _, v := range valuesToAdd {
		if _, ok := setMap[v]; !ok {
			set = append(unionSet, v)
			setMap[v] = true
		}
	}

	return unionSet
}

func ContainsWord(production, word string) bool {
	words := strings.Split(production, " ")

	for _, w := range words {
		if w == word {
			return true
		}
	}

	return false
}
