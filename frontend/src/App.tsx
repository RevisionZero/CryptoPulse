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
    const [connectionStatus, setConnectionStatus] = useState<'connected' | 'waiting' | 'disconnected'>('disconnected');
    const [showToast, setShowToast] = useState(false);
    const [toastMessage, setToastMessage] = useState('');

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

      // Connection opened
      socket.addEventListener('open', () => {
        console.log('Connected to WebSocket');
        setConnectionStatus('waiting');
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
          
          // Set to connected when we receive data
          if (symbolList.length > 0) {
            setConnectionStatus('connected');
          }
        } catch (error) {
          console.error('Error parsing WebSocket data:', error);
        }
      });

      // Handle errors
      socket.addEventListener('error', () => {
        console.error('WebSocket error');
        setConnectionStatus('disconnected');
      });

      // Connection closed
      socket.addEventListener('close', () => {
        console.log('Disconnected from WebSocket');
        setConnectionStatus('disconnected');
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

  const validateTicker = async (ticker: string): Promise<string> => {
    if (!ticker) return ticker; // Empty is valid (optional field)
    
    try {
      const response = await fetch(`https://api.binance.com/api/v3/exchangeInfo?symbol=${ticker}`);
      if (response.ok) {
        return ticker;
      }
      return 'INVALID_TICKER';
    } catch (error) {
      try {
        const newTicker = ticker + "USDT";
        const response = await fetch(`https://api.binance.com/api/v3/exchangeInfo?symbol=${newTicker}`);
        if (response.ok) {
          return newTicker;
        }
      } catch (err) {
        return 'INVALID_TICKER';
      }
      return 'INVALID_TICKER';
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
    const validatedTickers: string[] = [];
    let hasError = false;

    // Validate each non-empty ticker
    for (let i = 0; i < tickerInputs.length; i++) {
      const ticker = tickerInputs[i].trim();
      if (ticker) {
        const validatedTicker = await validateTicker(ticker);
        if (validatedTicker === 'INVALID_TICKER') {
          newErrors[i] = 'Invalid ticker';
          hasError = true;
        } else {
          validatedTickers.push(validatedTicker);
        }
      }
    }

    setValidationErrors(newErrors);
    setIsValidating(false);

    if (!hasError) {
      // Send validated tickers to WebSocket
      console.log('Sending tickers:', validatedTickers.toString());
      ws?.send(validatedTickers.toString());
      
      // Show toast notification
      setToastMessage(`✓ Tickers added: ${validatedTickers.join(', ')}`);
      setShowToast(true);
      setTimeout(() => setShowToast(false), 3000);
    }
  };

  return (
    <div className="app-container">
      {connectionStatus === 'connected' && (
        <div className="live-badge live-badge-green">
          <span className="live-dot"></span>
          LIVE DATA
        </div>
      )}
      {connectionStatus === 'waiting' && (
        <div className="live-badge live-badge-orange">
          <span className="live-dot"></span>
          WAITING FOR TICKERS
        </div>
      )}
      {connectionStatus === 'disconnected' && (
        <div className="live-badge live-badge-red">
          <span className="live-dot"></span>
          CONNECTION CLOSED
        </div>
      )}
      <h1>CryptoPulse</h1>
      <h2>Track cryptocurrency prices correlation in real-time</h2>
      
      {showToast && (
        <div className="toast">
          {toastMessage}
        </div>
      )}
      
      <div className="ticker-input-section">
        <h2>Select Tickers (2-5 required)</h2>
        <div className="ticker-inputs">
          {tickerInputs.map((ticker, index) => {
            const sampleTickers = ['BTC', 'ETH', 'SOL', 'XRP', 'DOGE'];
            return (
              <div key={index} className="ticker-input-group">
                <input
                  type="text"
                  placeholder={`${sampleTickers[index]}${index < 2 ? ' *' : ''}`}
                  value={ticker}
                  onChange={(e) => handleTickerChange(index, e.target.value)}
                  className={validationErrors[index] ? 'error' : ''}
                  disabled={isValidating}
                />
                {validationErrors[index] && (
                  <span className="error-message">{validationErrors[index]}</span>
                )}
              </div>
            );
          })}
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
        <>
          <div className="correlation-section">
            <div className="tooltip-container">
              <span className="help-icon">?</span>
              <div className="tooltip-text">
                Pearson Correlation (PCC) measures how closely two prices move together. +1.0 is perfect sync, -1.0 is perfect opposite
              </div>
            </div>
            <div className="correlation-header">
              <h2>Correlation Matrix</h2>
            </div>
          </div>
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
        </>
      ) : (
        <p>Waiting for data...</p>
      )}
      
      <footer className="app-footer">
        <p>© 2026 CryptoPulse | Developed By: Omar Shehada</p>
        <div className="footer-links">
          <a href="mailto:awadomar24@gmail.com">Email Me</a>
          <a href="https://github.com/RevisionZero" target="_blank" rel="noopener noreferrer" className="social-link">
            <svg height="20" width="20" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            GitHub
          </a>
          <a href="https://www.linkedin.com/in/omar-shehada-ch/" target="_blank" rel="noopener noreferrer" className="social-link">
            <svg height="20" width="20" viewBox="0 0 24 24" fill="currentColor">
              <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
            </svg>
            LinkedIn
          </a>
        </div>
      </footer>
    </div>
  );
}

export default App;
