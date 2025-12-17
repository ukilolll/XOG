import Cell from './Cell';

export default function Board({ cells, onCellClick }) {
  return (
    <div className="board">
      {cells.map((value, index) => (
        <Cell
          key={index}
          index={index}
          value={value}
          onClick={onCellClick}
        />
      ))}
    </div>
  );
}
