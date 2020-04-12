import * as React from 'react';

interface ToggleSetProps {
  toggle: {
    name: string;
    setting: string;
  };
  values: any;
  handleToggle: any;
}

const ToggleSet: React.FunctionalComponent<ToggleSetProps> = ({
  toggle,
  values,
  handleToggle,
}) => {
  return (
    <div className="toggle-set" key={toggle.setting}>
      <div className="settings-label">
        {toggle.name}{' '}
        <span className={'toggle-state'}>
          {values[toggle.setting] ? 'ON' : 'OFF'}
        </span>
      </div>
      <div
        onClick={e => handleToggle(e, toggle.setting)}
        className={values[toggle.setting] ? 'toggle active' : 'toggle inactive'}
      >
        <div className="switch"></div>
      </div>
    </div>
  );
};

export default ToggleSet;
