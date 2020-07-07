import * as React from 'react';
import WordSetToggle from '~/ui/wordset_toggle';

const CustomWords = ({
  words,
  onWordChange,
  selected,
  onToggle,
}) => {
  const [expanded, setExpanded] = React.useState(false);

  React.useEffect(() => {
    if (selected) {
      setExpanded(true);
    } else {
      setExpanded(false);
    }
  }, [selected]);

  const symbol = expanded ? '▾' : '▸';

  const wordCount = words
    .split(',')
    .map(w => w.trim())
    .filter(w => w.length > 0)
    .length;

  return (
    <div>
      <div className="btn-custom-word-set">
        <WordSetToggle
              key="custom"
              label={symbol + " Custom ("+wordCount+" words)"}
              selected={selected}
              onToggle={onToggle}>
        </WordSetToggle>
      </div>
    {expanded && (
      <div>
        <textarea value={words} aria-label="custom word set" onChange={(e) => onWordChange(e.target.value)} />
      </div>
    )}
  </div>
  );
};

export default CustomWords;
