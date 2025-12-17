export default function Cell({ index, value, onClick, disabled }) {
  return (
    <div
      className="cell"
      onClick={() => onClick(index)}
      data-index={index}
    >
      {value}
    </div>
  );
}
