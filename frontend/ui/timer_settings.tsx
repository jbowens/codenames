import * as React from 'react';
import ToggleSet from '~/ui/toggle-set';
import Toggle from '~/ui/toggle';

interface TimerSettingsProps {
  timer: [number, number];
  setTimer: (timer: [number, number]) => void;
  enforceTimerEnabled: boolean;
  setEnforceTimerEnabled: (newValue: boolean) => void;
}

const TimerSettings: React.FunctionalComponent<TimerSettingsProps> = ({
  timer,
  setTimer,
  enforceTimerEnabled,
  setEnforceTimerEnabled,
}) => {
  const [minutes, seconds] = timer || [];
  return (
    <div id="timer-settings">
      <ToggleSet
        toggle={{
          name: 'Timer',
          setting: 'timer',
          desc: 'If enabled a timer will countdown each team\'s turn.'
        }}
        values={{ timer }}
        handleToggle={() => {
          setTimer(!timer && [5, 0]);
        }}
      />
      {timer && (
        <div id="timer-duration">
          <div>
            <span>Duration:</span>
            <input
              type="number"
              name="minutes"
              id="minutes"
              min={0}
              max={59}
              value={minutes}
              onChange={(e) => {
                setTimer([parseInt(e?.target?.value), seconds]);
              }}
            />
            <label htmlFor="minutes">m</label>
            <input
              type="number"
              name="seconds"
              id="seconds"
              min={0}
              max={59}
              value={seconds}
              onChange={(e) => {
                setTimer([minutes, parseInt(e?.target?.value)]);
              }}
            />
            <label htmlFor="seconds">s</label>
          </div>
          <div>
            <span>Enforce timer:</span>
            <Toggle
              state={enforceTimerEnabled}
              handleToggle={() => setEnforceTimerEnabled(!enforceTimerEnabled)}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default TimerSettings;
