import { useEffect, useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

function App() {

    const [ws, setWs] = useState<WebSocket | null>(null);
    const [messages, setMessages] = useState<string[]>([]);

    useEffect(() => {
      // Create WebSocket connection
      const socket = new WebSocket('ws://localhost:8080');

      // Connection opened
      socket.addEventListener('open', (event) => {
        console.log('Connected to WebSocket');
        socket.send('Hello Server!');
      });

      // Listen for messages
      socket.addEventListener('message', (event) => {
        console.log('Message from server:', event.data);
        setMessages(prev => [...prev, event.data]);
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
    <>
      <div>

      </div>
    </>
  )
}

export default App
