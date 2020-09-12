import * as React from 'react';
import ToggleSet from '~/ui/toggle-set';

const settingToggles = [
  {
    name: 'Full-screen',
    setting: 'fullscreen',
    desc: 'Enlarge the board to take up the whole page.',
  },
  {
    name: 'Color-blind',
    setting: 'colorBlind',
    desc:
      'Add patterned borders to help color-blind players distinguish teams.',
  },
  {
    name: 'Dark',
    setting: 'darkMode',
    desc: 'Darken the mood.',
  },
  {
    name: 'Spymaster may guess',
    setting: 'spymasterMayGuess',
    desc: 'When enabled, clicking a word from spymaster view reveals the word.',
  },
];

export class Settings {
  static load() {
    try {
      const settingsBlob = localStorage.getItem('settings');
      return JSON.parse(settingsBlob) || {};
    } catch (e) {
      console.error(e);
      return {};
    }
  }

  static save(vals) {
    try {
      localStorage.setItem('settings', JSON.stringify(vals));
    } catch (e) {
      console.error(e);
    }
  }
}

export class SettingsButton extends React.Component {
  public handleClick(e) {
    e.preventDefault();
    this.props.onClick(e);
  }

  public render() {
    return (
      <button
        onClick={(e) => this.handleClick(e)}
        className="gear"
        aria-label="settings"
      >
        <svg
          width="30"
          height="30"
          viewBox="0 0 35 35"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M22.3344 4.86447L24.31 8.23766C21.9171 9.80387 21.1402 12.9586 22.5981 15.4479C23.038 16.1989 23.6332 16.8067 24.3204 17.2543L22.2714 20.7527C20.6682 19.9354 18.6888 19.9151 17.0088 20.8712C15.3443 21.8185 14.3731 23.4973 14.2734 25.2596H10.3693C10.3241 24.4368 10.087 23.612 9.64099 22.8504C8.16283 20.3266 4.93593 19.4239 2.34593 20.7661L0.342913 17.3461C2.85907 15.8175 3.70246 12.5796 2.21287 10.0362C1.74415 9.23595 1.09909 8.59835 0.354399 8.14386L2.34677 4.74208C3.95677 5.5788 5.95446 5.60726 7.64791 4.64346C9.31398 3.69524 10.2854 2.0141 10.3836 0.25H14.267C14.2917 1.11932 14.5297 1.99505 15.0012 2.80013C16.4866 5.33635 19.738 6.23549 22.3344 4.86447ZM15.0038 17.3703C17.6265 15.8776 18.5279 12.5685 17.0114 9.97937C15.4963 7.39236 12.1437 6.50866 9.52304 8.00013C6.90036 9.4928 5.99896 12.8019 7.5154 15.391C9.03058 17.978 12.3832 18.8617 15.0038 17.3703Z"
            transform="translate(12.7548) rotate(30)"
            fill="#EEE"
            stroke="#BBB"
            strokeWidth="0.5"
          />
        </svg>
      </button>
    );
  }
}

export class SettingsPanel extends React.Component {
  public render() {
    return (
      <div className="settings">
        <div
          onClick={(e) => this.props.toggleView(e)}
          className="close-settings"
        >
          <svg
            width="32"
            height="32"
            viewBox="0 0 32 32"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M0 0L30 30M30 0L0 30"
              transform="translate(1 1)"
              stroke="black"
              strokeWidth="2"
              role="button"
              aria-label="close settings"
            />
          </svg>
        </div>
        <div className="settings-content">
          <h2>SETTINGS</h2>
          <div className="toggles">
            {settingToggles.map((toggle) => (
              <ToggleSet
                key={toggle.name}
                values={this.props.values}
                toggle={toggle}
                handleToggle={this.props.toggle}
              />
            ))}
          </div>
        </div>
      </div>
    );
  }
}
