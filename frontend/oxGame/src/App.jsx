import { useEffect, useState } from 'react';
import './App.css';
import Board from './components/Board';
import { useWebSocket } from './hooks/useWebSocket';

function App() {
  const { isConnected, lastMessage, send } = useWebSocket('/api/apiv1');
  
  const [cells, setCells] = useState(Array(9).fill(''));
  const [myRole, setMyRole] = useState('');
  const [myTurn, setMyTurn] = useState(false);
  const [status, setStatus] = useState('Connecting...');
  const [result, setResult] = useState('');
  const [gameEnded, setGameEnded] = useState(false);

  // Handle WebSocket messages
  useEffect(() => {
    if (!lastMessage) return;

    switch (lastMessage.type) {
      case 'wait':
        setStatus('Waiting for another player...');
        break;

      case 'start':
        setStatus('Game started!');
        break;

      case 'role':
        setMyRole(lastMessage.data);
        break;

      case 'turn': {
        const turn = lastMessage.data;
        const isTurn = turn === myRole;
        setMyTurn(isTurn);
        setStatus(isTurn ? 'Your turn!' : "Opponent's turn");
        break;
      }

      case 'play': {
        const { index, value } = lastMessage.data;
        setCells(prev => {
          const newCells = [...prev];
          newCells[index] = value;
          return newCells;
        });
        break;
      }

      case 'disconnect': {
        const data = lastMessage.data;
        if (data === myRole) {
          setResult('Time Out disconnected.');
        } else {
          setResult('You win!!! Opponent disconnect.');
        }
        setGameEnded(true);
        break;
      }

      case 'end': {
        const gameResult = lastMessage.data;
        if (gameResult === 'draw') {
          setResult("It's a draw!");
        } else {
          setResult(gameResult === myRole ? 'You win!' : 'You lose!');
        }
        setMyTurn(false);
        setGameEnded(true);
        break;
      }

      default:
        break;
    }
  }, [lastMessage, myRole]);

  const handleCellClick = (index) => {
    if (myTurn && cells[index] === '') {
      send({
        type: 'play',
        data: index
      });
      setMyTurn(false);
    }
  };

  const handlePlayAgain = () => {
    setCells(Array(9).fill(''));
    setResult('');
    setGameEnded(false);
    setStatus('Waiting for another player...');
  };

  return (
    <div className="app-container">
      <h1>XO Game</h1>
      {myRole && <div className="role">You are {myRole.toUpperCase()}</div>}
      <Board cells={cells} onCellClick={handleCellClick} />
      <div className="status">{status}</div>
      {result && <div className="result">{result}</div>}
      {gameEnded && (
        <div className="button-container">
          <button onClick={handlePlayAgain}>Play again</button>
        </div>
      )}
    </div>
  );
}

export default App;
