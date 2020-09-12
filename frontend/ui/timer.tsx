import * as React from 'react';

function getTimeRemaining(endTime: number) {
  const diff = endTime - Date.now();
  const seconds = Math.max(Math.floor((diff / 1000) % 60), 0);
  const minutes = Math.max(Math.floor((diff / 1000 / 60) % 60), 0);
  return {
    total: Math.floor(diff / 1000),
    minutes: `${minutes < 10 ? '0' : ''}${minutes}`,
    seconds: `${seconds < 10 ? '0' : ''}${seconds}`,
  };
}

interface TimerProps {
  roundStartedAt: number;
  timerDurationMs: number;
  handleExpiration: () => void;
  freezeTimer: boolean;
}

const Timer: React.FunctionComponent<TimerProps> = ({
  roundStartedAt,
  timerDurationMs,
  handleExpiration,
  freezeTimer = false,
}) => {
  const [timeRemaining, setTimeRemaining] = React.useState(undefined);
  const endTime = new Date(roundStartedAt).getTime() + timerDurationMs + 1000;

  React.useEffect(() => {
    const timeRemaining = getTimeRemaining(endTime - 1000);
    if (timeRemaining.total < 0) {
      handleExpiration();
    }
    const timeout = freezeTimer
      ? null
      : setTimeout(() => setTimeRemaining(timeRemaining), 1000);

    return () => {
      clearTimeout(timeout);
    };
  }, [timeRemaining]);

  React.useEffect(() => {
    setTimeRemaining(getTimeRemaining(endTime));
  }, [endTime]);

  if (!timeRemaining?.total && timeRemaining?.total !== 0) return null;

  let color;
  if (timeRemaining.total <= 30) color = '#F70';
  if (timeRemaining.total <= 10) color = '#E22';
  return (
    <span
      style={{ color }}
      role="img"
      aria-label={
        'Time remaining: ' +
        timeRemaining.minutes.toString() +
        ':' +
        timeRemaining.seconds.toString()
      }
    >
      {timeRemaining.minutes}:{timeRemaining.seconds}
    </span>
  );
};

export default Timer;
