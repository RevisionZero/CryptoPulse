import { useEffect, useState } from 'react'
import './App.css'

type CorrelationMatrix = {
  [symbol: string]: {
    [symbol: string]: number;
  };
};

function App() {
    const [ws, setWs] = useState<WebSocket | null>(null);
    const [correlationMatrix, setCorrelationMatrix] = useState<CorrelationMatrix>({});
    const [symbols, setSymbols] = useState<string[]>([]);

    useEffect(() => {
      // Create WebSocket connection
      const socket = new WebSocket('ws://localhost:8080/ws');

      // Connection opened
      socket.addEventListener('open', (event) => {
        console.log('Connected to WebSocket');
      });

      // Listen for messages
      socket.addEventListener('message', (event) => {
        console.log('Message from server ', event.data);
        try {
          const data: CorrelationMatrix = JSON.parse(event.data);
          setCorrelationMatrix(data);
          
          // Extract unique symbols from the matrix
          const symbolList = Object.keys(data);
          setSymbols(symbolList);
        } catch (error) {
          console.error('Error parsing WebSocket data:', error);
        }
      });

      // Handle errors
      socket.addEventListener('error', (event) => {
        console.error('WebSocket error:', event);
      });

      // Connection closed
      socket.addEventListener('close', (event) => {
        console.log('Disconnected from WebSocket');
      });

      setWs(socket);

      // Cleanup on unmount
      return () => {
        socket.close();
      };
    }, []);

  return (
    <div className="app-container">
      <h1>NexusCorr - Correlation Matrix</h1>
      
      {symbols.length > 0 ? (
        <div className="table-container">
          <table className="correlation-table">
            <thead>
              <tr>
                <th></th>
                {symbols.map(symbol => (
                  <th key={symbol}>{symbol}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {symbols.map(rowSymbol => (
                <tr key={rowSymbol}>
                  <th>{rowSymbol}</th>
                  {symbols.map(colSymbol => {
                    const value = correlationMatrix[rowSymbol]?.[colSymbol];
                    const displayValue = value !== undefined ? value.toFixed(4) : '-';
                    
                    return (
                      <td 
                        key={colSymbol}
                        className={rowSymbol === colSymbol ? 'diagonal' : ''}
                        style={{
                          backgroundColor: value !== undefined 
                            ? `rgba(${value >= 0 ? '0, 255, 0' : '255, 0, 0'}, ${Math.abs(value) * 0.3})`
                            : 'transparent'
                        }}
                      >
                        {displayValue}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p>Waiting for data...</p>
      )}
    </div>
  )
}

export default App
