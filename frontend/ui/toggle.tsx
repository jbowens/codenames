import * as React from 'react';

interface ToggleProps {
  state: boolean;
  handleToggle: any;
}

const Toggle: React.FunctionalComponent<ToggleProps> = ({
  state,
  handleToggle,
}) => {
  return (
    <div
      onClick={handleToggle}
      className={state ? 'toggle active' : 'toggle inactive'}
    >
      <div className="switch"></div>
    </div>
  );
};

export default Toggle;
