import * as React from 'react';
import { OriginalWords, WordsPicker } from '~/ui/words_picker';

// TODO: remove jquery dependency
// https://stackoverflow.com/questions/47968529/how-do-i-use-jquery-and-jquery-ui-with-parcel-bundler
var jquery = require('jquery');
window.$ = window.jQuery = jquery;

export class Lobby extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      newGameName: this.props.defaultGameID,
      selectedGame: null,
      words: OriginalWords,
    };
  }

  public newGameTextChange(e) {
    this.setState({ newGameName: e.target.value });
  }

  public handleNewGame(e) {
    e.preventDefault();
    if (!this.state.newGameName) {
      return;
    }

    const gameID = this.state.newGameName;
    $.post(
      '/next-game',
      JSON.stringify({
        game_id: gameID,
        word_set: this.state.words.split(', '),
      }),
      g => {
        const newURL = (document.location.pathname = '/' + gameID);
        window.location = newURL;
      }
    );
  }

  public onChangedWords(words) {
    this.setState({ words: words });
  }

  public render() {
    return (
      <div id="lobby">
        <div id="available-games">
          <form id="new-game">
            <p className="intro">
              Play Codenames online across multiple devices on a shared board.
              To create a new game or join an existing game, enter a game
              identifier and click 'GO'.
            </p>
            <input
              type="text"
              id="game-name"
              autoFocus
              onChange={this.newGameTextChange.bind(this)}
              value={this.state.newGameName}
            />
            <button onClick={this.handleNewGame.bind(this)}>Go</button>

            <div id="new-game-options">
              <WordsPicker
                words={this.state.words}
                onChange={this.onChangedWords.bind(this)}
              />
            </div>
          </form>
        </div>
      </div>
    );
  }
}
