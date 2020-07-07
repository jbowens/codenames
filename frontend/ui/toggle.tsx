import * as React from 'react';

interface ToggleProps {
  name: string;
  state: boolean;
  handleToggle: any;
}

const Toggle: React.FunctionalComponent<ToggleProps> = ({
  name,
  state,
  handleToggle,
}) => {
  return (
    <div
      onClick={handleToggle}
      className={state ? 'toggle active' : 'toggle inactive'}
    >
      <div 
        className="switch"
        role="button"
        aria-label={name}
        aria-pressed={!!state}
      ></div>
    </div>
  );
};

export default Toggle;
