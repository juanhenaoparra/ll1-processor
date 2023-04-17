import React, { useState } from "react";
import { postLL1 } from "./api/api.js";
import MyTable from "./components/table.js";

function App() {
  const [text, setText] = useState("");
  const [result, setResult] = useState({});
  const [grammar, setGrammar] = useState({});

  function formatGrammar(input) {
    // Dividir la entrada en líneas
    const lines = input.split("\n");

    // Iterar a través de las líneas y dividir cada línea en tokens
    const grammar = { productions_set: {} };
    lines.forEach(line => {
      const tokens = line.trim().split(/\s*->\s*|\s*\|\s*/);
      const nonTerminal = tokens[0];
      const productions = tokens.slice(1);
      if (!grammar.productions_set[nonTerminal]) {
        grammar.productions_set[nonTerminal] = [];
      }
      productions.forEach(p => {
        grammar.productions_set[nonTerminal].push(p);
      });
    });

    return grammar;
  }
  function getNonTerminalOrder(grammarString) {
    const productions = grammarString.split('\n');
    const order = [];

    for (const production of productions) {
      const [left] = production.split(' -> ');
      if (!order.includes(left)) {
        order.push(left);
      }
    }
    const order1 = { order: order }
    return order1;
  }

  const grammarText = `AL -> id := p
P -> P or D | D
D -> D and C | C
C -> S | not (P)
S -> (P) | OP REL OP | true | false
REL -> = | < | <= | > | >= | <>
OP -> id | num`;

  const handleTextChange = (event) => {
    setText(event.target.value);
  };

  const handleButtonClick = (event) => {
    const grammar = formatGrammar(text);
    const order = getNonTerminalOrder(text);
    const combinedGrammar = Object.assign({}, order, grammar);
    const jsonGrammar = JSON.stringify(combinedGrammar);
    postLL1("http://localhost:3002/ll1", jsonGrammar)
      .then((response) => {
        if (response.result) {
          setResult(response.result)
          setGrammar(response.grammar)
        }
        console.log(response);
      })
      .catch((error) => {
        console.error(error);
      });
  };
  console.log(result, grammar);

  return (
    <div>
      <h2>Ingrese la Gramática</h2>
      <textarea
        style={{
          width: "100%",
          height: "200px",
          padding: "10px",
          boxSizing: "border-box",
          fontFamily: "monospace",
        }}
        onChange={handleTextChange}
      />
      <button
        style={{
          padding: "10px",
          backgroundColor: "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
          marginTop: "10px",
        }}
        onClick={handleButtonClick}
      >
        Enviar
      </button>
      {Object.keys(result)
        ? (<MyTable
          grammar={grammar}
          result= {result}
        />)
        : null}
    </div>
  );
}

export default App;
