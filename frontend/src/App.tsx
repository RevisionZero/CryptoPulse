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
    const [tickerInputs, setTickerInputs] = useState<string[]>(['', '', '', '', '']);
    const [validationErrors, setValidationErrors] = useState<string[]>(['', '', '', '', '']);
    const [isValidating, setIsValidating] = useState(false);

    useEffect(() => {
      // Create WebSocket connection
      // const socket = new WebSocket('ws://localhost:8080/ws');

      // Connection opened
      // socket.addEventListener('open', (event) => {
      //   console.log('Connected to WebSocket');
      // });

      // 1. Determine if we should use secure 'wss' or standard 'ws'
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      
      // 2. Build the URL using the current domain (window.location.host)
      // This automatically handles localhost:5173 (dev) or cryptopulseapp.dev (prod)
      const wsUrl = `${protocol}//${window.location.host}/ws`;

      // 3. Create the WebSocket connection
      const socket = new WebSocket(wsUrl);

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

  const handleTickerChange = (index: number, value: string) => {
    const newInputs = [...tickerInputs];
    newInputs[index] = value.toUpperCase();
    setTickerInputs(newInputs);
    
    // Clear validation error when user types
    const newErrors = [...validationErrors];
    newErrors[index] = '';
    setValidationErrors(newErrors);
  };

  const validateTicker = async (ticker: string): Promise<boolean> => {
    if (!ticker) return true; // Empty is valid (optional field)
    
    try {
      const response = await fetch(`https://api.binance.com/api/v3/exchangeInfo?symbol=${ticker}`);
      if (response.ok) {
        return true;
      }
      return false;
    } catch (error) {
      return false;
    }
  };

  const handleValidateAndSubmit = async () => {
    // Filter out empty tickers
    const nonEmptyTickers = tickerInputs.filter(t => t.trim() !== '');
    
    // Check minimum 2 tickers
    if (nonEmptyTickers.length < 2) {
      alert('Please enter at least 2 tickers');
      return;
    }

    setIsValidating(true);
    const newErrors = ['', '', '', '', ''];
    let hasError = false;

    // Validate each non-empty ticker
    for (let i = 0; i < tickerInputs.length; i++) {
      const ticker = tickerInputs[i].trim();
      if (ticker) {
        const isValid = await validateTicker(ticker);
        if (!isValid) {
          newErrors[i] = 'Invalid ticker';
          hasError = true;
        }
      }
    }

    setValidationErrors(newErrors);
    setIsValidating(false);

    if (!hasError) {
      console.log('Valid tickers:', nonEmptyTickers.toString());
      // TODO: Send to WebSocket
      alert(`Valid tickers ready to send: ${nonEmptyTickers.join(',')}`);
      ws?.send(nonEmptyTickers.toString()); 
    }
  };

  return (
    <div className="app-container">
      <h1>NexusCorr - Correlation Matrix</h1>
      
      <div className="ticker-input-section">
        <h2>Select Tickers (2-5 required)</h2>
        <div className="ticker-inputs">
          {tickerInputs.map((ticker, index) => (
            <div key={index} className="ticker-input-group">
              <input
                type="text"
                placeholder={`Ticker ${index + 1}${index < 2 ? ' *' : ''}`}
                value={ticker}
                onChange={(e) => handleTickerChange(index, e.target.value)}
                className={validationErrors[index] ? 'error' : ''}
                disabled={isValidating}
              />
              {validationErrors[index] && (
                <span className="error-message">{validationErrors[index]}</span>
              )}
            </div>
          ))}
        </div>
        <button 
          onClick={handleValidateAndSubmit}
          disabled={isValidating}
          className="validate-button"
        >
          {isValidating ? 'Validating...' : 'Validate & Add Tickers'}
        </button>
      </div>
      
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
  );
}

export default App;
