import React from "react";


function MyTable({ grammar, result }) {
    if (!grammar || !grammar.order || (!result && !result.first)) {
        return <p>No hay datos para mostrar</p>;
    }
    return (
        <>
            <table style={{ borderCollapse: "collapse", width: "30%", marginBottom: "1rem", border: "1px solid black", marginTop: "1rem" }}>
                <thead>
                    <tr>
                        <th style={{ border: "1px solid black", padding: "0.5rem" }}>No terminales</th>
                        <th style={{ border: "1px solid black", padding: "0.5rem" }}>Primeros</th>
                    </tr>
                </thead>
                <tbody>
                    {Object.keys(result.first).map((key) => (
                        <tr key={key}>
                            <td style={{ border: "1px solid black", padding: "0.5rem" }}>{key}</td>
                            <td style={{ border: "1px solid black", padding: "0.5rem" }}>
                                <ul className="list-outside list-disc">
                                    {result.first[key].map((item, idx) => (
                                        <React.Fragment key={idx}>{`${item}, `}</React.Fragment>
                                    ))}
                                </ul>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
            {
                result.follow && (
                    <table style={{ borderCollapse: "collapse", width: "30%", marginBottom: "1rem", border: "1px solid black", marginTop: "1rem" }}>
                        <thead>
                            <tr>
                                <th style={{ border: "1px solid black", padding: "0.5rem" }}>No terminales</th>
                                <th style={{ border: "1px solid black", padding: "0.5rem" }}>siguientes</th>
                            </tr>
                        </thead>
                        <tbody>
                            {Object.keys(result.follow).map((key) => (
                                <tr key={key}>
                                    <td style={{ border: "1px solid black", padding: "0.5rem" }}>{key}</td>
                                    <td style={{ border: "1px solid black", padding: "0.5rem" }}>
                                        <ul className="list-outside list-disc">
                                            {result.follow[key].map((item, idx) => (
                                                <React.Fragment key={idx}>{`${item}, `}</React.Fragment>
                                            ))}
                                        </ul>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )
            }
            {
                result.prediction && (
                    <table style={{ borderCollapse: "collapse", width: "30%", marginBottom: "1rem", border: "1px solid black", marginTop: "1rem" }}>
                        <thead>
                            <tr>
                                <th style={{ border: "1px solid black", padding: "0.5rem" }}>No terminales</th>
                                <th style={{ border: "1px solid black", padding: "0.5rem" }}>conjunto predicción</th>
                            </tr>
                        </thead>
                        <tbody>
                            {Object.keys(result.prediction).map((key) => (
                                <tr key={key}>
                                    <td style={{ border: "1px solid black", padding: "0.5rem" }}>{key}</td>
                                    <td style={{ border: "1px solid black", padding: "0.5rem" }}>
                                        <ul className="list-outside list-disc">
                                            {result.prediction[key].map((item, idx) => (
                                                <React.Fragment key={idx}>{`${item}, `}</React.Fragment>
                                            ))}
                                        </ul>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )
            }
        </>

    );
}

export default MyTable;
