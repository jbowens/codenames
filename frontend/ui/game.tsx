import * as React from 'react';
import { Settings, SettingsButton, SettingsPanel } from '~/ui/settings';

// TODO: remove jquery dependency
// https://stackoverflow.com/questions/47968529/how-do-i-use-jquery-and-jquery-ui-with-parcel-bundler
let jquery = require('jquery');
window.$ = window.jQuery = jquery;

const defaultFavicon = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAA8SURBVHgB7dHBDQAgCAPA1oVkBWdzPR84kW4AD0LCg36bXJqUcLL2eVY/EEwDFQBeEfPnqUpkLmigAvABK38Grs5TfaMAAAAASUVORK5CYII=';
const blueTurnFavicon = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAmSURBVHgB7cxBAQAABATBo5ls6ulEiPt47ASYqJ6VIWUiICD4Ehyi7wKv/xtOewAAAABJRU5ErkJggg==';
const redTurnFavicon = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAmSURBVHgB7cwxAQAACMOwgaL5d4EiELGHoxGQGnsVaIUICAi+BAci2gJQFUhklQAAAABJRU5ErkJggg==';
export class Game extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      game: null,
      mounted: true,
      settings: Settings.load(),
      mode: 'game',
      codemaster: false,
    };
  }

  public extraClasses() {
    var classes = '';
    if (this.state.settings.colorBlind) {
      classes += ' color-blind';
    }
    if (this.state.settings.darkMode) {
      classes += ' dark-mode';
    }
    if (this.state.settings.fullscreen) {
      classes += ' full-screen';
    }
    return classes;
  }

  public handleKeyDown(e) {
    if (e.keyCode == 27) {
      this.setState({ mode: 'game' });
    }
  }

  public componentDidMount(prevProps, prevState) {
    window.addEventListener('keydown', this.handleKeyDown.bind(this));
    this.setDarkMode(prevProps, prevState);
    this.setTurnIndicatorFavicon(prevProps, prevState);
    this.refresh();
  }

  public componentWillUnmount() {
    window.removeEventListener('keydown', this.handleKeyDown.bind(this));
    document.getElementById("favicon").setAttribute("href", defaultFavicon);
    this.setState({ mounted: false });
  }

  public componentDidUpdate(prevProps, prevState) {
    this.setDarkMode(prevProps, prevState);
    this.setTurnIndicatorFavicon(prevProps, prevState);
  }

  private setDarkMode(prevProps, prevState) {
    if (!prevState?.settings.darkMode && this.state.settings.darkMode) {
      document.body.classList.add('dark-mode');
    }
    if (prevState?.settings.darkMode && !this.state.settings.darkMode) {
      document.body.classList.remove('dark-mode');
    }
  }

  private setTurnIndicatorFavicon(prevProps, prevState) {
    if (
      prevState?.game?.winning_team !== this.state.game?.winning_team ||
      prevState?.game?.round !== this.state.game?.round ||
      prevState?.game?.state_id !== this.state.game?.state_id
    ) {
      if (this.state.game?.winning_team) {
        document.getElementById("favicon").setAttribute("href", defaultFavicon);
      } else {
        document.getElementById("favicon").setAttribute("href", this.currentTeam() === 'blue' ? blueTurnFavicon : redTurnFavicon);
      }
    }
  }

  public refresh() {
    if (!this.state.mounted) {
      return;
    }

    let state_id = "";
    if (this.state.game && this.state.game.state_id) {
      state_id = this.state.game.state_id;
    }

    const body = { game_id: this.props.gameID, state_id: state_id };
    $.ajax({
      url: '/game-state',
      type: 'POST',
      data: JSON.stringify(body),
      contentType:'application/json; charset=utf-8',
      dataType: 'json',
      success: (data => {
        if (this.state.game && data.created_at != this.state.game.created_at) {
          this.setState({ codemaster: false });
        }
        this.setState({ game: data });
      }),
      complete: () => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      },
    });
  }

  public toggleRole(e, role) {
    e.preventDefault();
    this.setState({ codemaster: role == 'codemaster' });
  }

  public guess(e, idx, word) {
    e.preventDefault();
    if (this.state.game.revealed[idx]) {
      return; // ignore if already revealed
    }
    if (this.state.game.winning_team) {
      return; // ignore if game is over
    }
    $.post(
      '/guess',
      JSON.stringify({
        game_id: this.state.game.id,
        index: idx,
      }),
      g => {
        this.setState({ game: g });
      }
    );
  }

  public currentTeam() {
    if (this.state.game.round % 2 == 0) {
      return this.state.game.starting_team;
    }
    return this.state.game.starting_team == 'red' ? 'blue' : 'red';
  }

  public remaining(color) {
    var count = 0;
    for (var i = 0; i < this.state.game.revealed.length; i++) {
      if (this.state.game.revealed[i]) {
        continue;
      }
      if (this.state.game.layout[i] == color) {
        count++;
      }
    }
    return count;
  }

  public endTurn() {
    $.post(
      '/end-turn',
      JSON.stringify({ game_id: this.state.game.id }),
      g => {
        this.setState({ game: g });
      }
    );
  }

  public nextGame(e) {
    e.preventDefault();
    // Ask for confirmation when current game hasn't finished
    let allowNextGame = (
      this.state.game.winning_team ||
      confirm("Do you really want to start a new game?")
    );
    if (!allowNextGame) {
      return;
    }
    $.post(
      '/next-game',
      JSON.stringify({
        game_id: this.state.game.id,
        word_set: this.state.game.word_set,
        create_new: true,
      }),
      g => {
        this.setState({ game: g, codemaster: false });
      }
    );
  }

  public toggleSettingsView(e) {
    if (e != null) {
      e.preventDefault();
    }
    if (this.state.mode == 'settings') {
      this.setState({ mode: 'game' });
    } else {
      this.setState({ mode: 'settings' });
    }
  }

  public toggleSetting(e, setting) {
    if (e != null) {
      e.preventDefault();
    }
    const vals = { ...this.state.settings };
    vals[setting] = !vals[setting];
    this.setState({ settings: vals });
    Settings.save(vals);
  }

  render() {
    if (!this.state.game) {
      return <p className="loading">Loading&hellip;</p>;
    }
    if (this.state.mode == 'settings') {
      return (
        <SettingsPanel
          toggleView={e => this.toggleSettingsView(e)}
          toggle={(e, setting) => this.toggleSetting(e, setting)}
          values={this.state.settings}
        />
      );
    }

    let status, statusClass;
    if (this.state.game.winning_team) {
      statusClass = this.state.game.winning_team + ' win';
      status = this.state.game.winning_team + ' wins!';
    } else {
      statusClass = this.currentTeam();
      status = this.currentTeam() + "'s turn";
    }

    let endTurnButton;
    if (!this.state.game.winning_team && !this.state.codemaster) {
      endTurnButton = (
        <div id="end-turn-cont">
          <button onClick={e => this.endTurn(e)} id="end-turn-btn">
            End {this.currentTeam()}&#39;s turn
          </button>
        </div>
      );
    }

    let otherTeam = 'blue';
    if (this.state.game.starting_team == 'blue') {
      otherTeam = 'red';
    }

    let shareLink = null;
    if (!this.state.settings.fullscreen) {
      shareLink = (
        <div id="share">
          Send this link to friends:
          <a className="url" href={window.location.href}>
            {window.location.href}
          </a>
        </div>
      );
    }

    return (
      <div
        id="game-view"
        className={
          (this.state.codemaster ? 'codemaster' : 'player') +
          this.extraClasses()
        }
      >
        {shareLink}
        <div id="status-line" className={statusClass}>
          <div id="remaining">
            <span className={this.state.game.starting_team + '-remaining'}>
              {this.remaining(this.state.game.starting_team)}
            </span>
            &nbsp;&ndash;&nbsp;
            <span className={otherTeam + '-remaining'}>
              {this.remaining(otherTeam)}
            </span>
          </div>
          <div id="status" className="status-text">
            {status}
          </div>
          {endTurnButton}
        </div>
        <div className="board">
          {this.state.game.words.map((w, idx) => (
            <div
              key={idx}
              className={
                'cell ' +
                this.state.game.layout[idx] +
                ' ' +
                (this.state.game.revealed[idx] ? 'revealed' : 'hidden-word')
              }
              onClick={e => this.guess(e, idx, w)}
            >
              <span className="word">{w}</span>
            </div>
          ))}
        </div>
        <form
          id="mode-toggle"
          className={
            this.state.codemaster ? 'codemaster-selected' : 'player-selected'
          }
        >
          <SettingsButton
            onClick={e => {
              this.toggleSettingsView(e);
            }}
          />
          <button
            onClick={e => this.toggleRole(e, 'player')}
            className="player"
          >
            Player
          </button>
          <button
            onClick={e => this.toggleRole(e, 'codemaster')}
            className="codemaster"
          >
            Spymaster
          </button>
          <button onClick={e => this.nextGame(e)} id="next-game-btn">
            Next game
          </button>
        </form>
        <div id="coffee"><a href="https://www.buymeacoffee.com/jbowens" target="_blank">Buy the developer a coffee.</a></div>
      </div>
    );
  }
}
