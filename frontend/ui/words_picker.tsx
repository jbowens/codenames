import * as React from 'react';
import OriginalWords from '~/words.json';

const WordsPicker = ({
  words,
  onWordChange,
  language,
  selectedLanguage,
  onSelectedLanguageChange,
}) => {
  const [expanded, setExpanded] = React.useState(false);

  React.useEffect(() => {
    if (language !== selectedLanguage) {
      setExpanded(false);
    }
  }, [language, selectedLanguage]);

  const symbol = expanded ? '▾' : '▸';
  const label = words === OriginalWords[language] ? language : 'Custom';

  const wordsArray = words
    .trim()
    .split(',')
    .map(w => w.trim())
    .filter(w => w.length > 0);

  let warning = null;
  if (wordsArray.length < 25) {
    warning = <div className="warning">must have 25+ words</div>;
  }

  return (
    <div>
      <label className="language-group">
        <input
          type="radio"
          value={language}
          checked={language === selectedLanguage}
          onChange={onSelectedLanguageChange}
        />
        <div
          className="btn-configured-word-set"
          onClick={() => {
            setExpanded(!expanded);
          }}
        >
          <div className="symbol">{symbol}</div>
          <div className="label">Words: {label}</div>
        </div>
        {warning}
      </label>
      {expanded && (
        <div>
          <textarea value={words} onChange={onWordChange} />
        </div>
      )}
    </div>
  );
};

export default WordsPicker;
