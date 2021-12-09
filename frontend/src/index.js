import React from 'react';
import ReactDOM from 'react-dom';

function App() {
    const [data, setData] = React.useState(null);

    React.useEffect(() => {
        fetch("/api")
        .then((res) => res.text())
        .then((data) => setData(data));
    }, []);

    return (
        <div>
            <header>
                <h1>FWENDS</h1>
            </header>
            <main>
                <p>{!data ? "Loading..." : data}</p>
            </main>
        </div>
    );
}

ReactDOM.render(
    <App />,
    document.getElementById('root')
);
