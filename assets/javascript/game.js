window.Game = React.createClass({
    propTypes: {
        gameID: React.PropTypes.string,
    },

    getInitialState: function() {
        return {
            game: null,
            mounted: true,
            codemaster: false,
        };
    },

    componentWillMount: function() {
      this.refresh();
    },

    componentWillUnmount: function() {
      this.setState({mounted: false});
    },

    refresh: function() {
      if (!this.state.mounted) {
          return;
      }

      $.get('/game/' + this.props.gameID, (data) => {
          if (this.state.game && data.round != this.state.game.round) {
            this.changedTurn();
          }
          this.setState({game: data});
          setTimeout(this.refresh, 3000);
      });
    },

    changedTurn: function() {
        $('#clue-word')[0].value = "";
        $('#clue-word')[0].disabled = false;
        $('#clue-count')[0].value = "";
        $('#clue-count')[0].disabled = false;
    },

    toggleRole: function(e, role) {
        e.preventDefault();
        this.setState({codemaster: role=='codemaster'});
    },

    guess: function(e, idx, word) {
        e.preventDefault();
        if (this.state.codemaster) {
            return; // ignore if codemaster view
        }
        if (this.state.game.revealed[idx]) {
            return; // ignore if already revealed
        }
        if (this.state.game.winning_team) {
            return; // ignore if game is over
        }
        $.post('/guess', JSON.stringify({
            game_id: this.state.game.id,
            index: idx,
        }), (g) => { this.setState({game: g}); });
    },

    currentTeam: function() {
        if (this.state.game.round % 2 == 0) {
            return this.state.game.starting_team;
        }
        return this.state.game.starting_team == 'red' ? 'blue' : 'red';
    },

    remaining: function(color) {
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
    },

    endTurn: function() {
        $.post('/end-turn', JSON.stringify({game_id: this.state.game.id}),
              (g) => { this.setState({game: g}); this.changedTurn(); });
    },

    nextGame: function(e) {
        e.preventDefault();
        $.post('/next-game', JSON.stringify({game_id: this.state.game.id}),
              (g) => { this.setState({game: g}); });
    },

    giveClue: function(e) {
        let clueWord = $('#clue-word')[0].value;
        let clueCount = parseInt($('#clue-count')[0].value);
        // How does this not have an error callback :(
        $.post('/clue', JSON.stringify({
            game_id: this.state.game.id,
            word: clueWord,
            count: clueCount
        }), (g) => {
            this.setState({game: g});
            $('#clue-word')[0].disabled = true;
            $('#clue-count')[0].disabled = true;
        });
    },

    render: function() {
        if (!this.state.game) {
            return (<p className="loading">Loading&hellip;</p>);
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
            endTurnButton = (<button onClick={(e) => this.endTurn(e)} id="end-turn-btn">End {this.currentTeam()}&#39;s turn</button>)
        }

        let otherTeam = 'blue';
        if (this.state.game.starting_team == 'blue') {
            otherTeam = 'red';
        }

        let clueDOM;
        if (this.state.codemaster) {
            clueDOM = (
                <div className="clue-line">
                    <input type="text" id="clue-word" placeholder="word" autoFocus />
                    <input type="number" id="clue-count" placeholder="count" />
                    <button onClick={this.giveClue}>Give Clue</button>
                </div>
            )
        } else {
            clueDOM = (
                <div className="clue-line">
                    {this.state.game.clue ? "Clue: " + this.state.game.clue.word + "(" + this.state.game.clue.count + ")" : "Waiting for clue..."}
                </div>
            );
        }

        return (
            <div id="game-view" className={this.state.codemaster ? "codemaster" : "player"}>
                <div id="share">
                  Send this link to friends: <a className="url" href={window.location.href}>{window.location.href}</a>
                </div>
                <div id="status-line" className={statusClass}>
                    <div id="status" className="status-text">{status}</div>
                </div>
                {clueDOM}
                <div id="button-line">
                    <div id="remaining">
                        <span className={this.state.game.starting_team+"-remaining"}>{this.remaining(this.state.game.starting_team)}</span>
                        &nbsp;&ndash;&nbsp;
                        <span className={otherTeam + "-remaining"}>{this.remaining(otherTeam)}</span>
                    </div>
                    {endTurnButton}
                    <div className="clear"></div>
                </div>
                <div id="board">
                  {this.state.game.words.map((w, idx) =>
                    (
                        <div for={idx}
                             className={"cell " + this.state.game.layout[idx] + " " + (this.state.game.revealed[idx] ? "revealed" : "hidden")}
                             onClick={(e) => this.guess(e, idx, w)}
                        >
                            <span className="word">{w}</span>
                        </div>
                    )
                  )}
                </div>
                <form id="mode-toggle" className={this.state.codemaster ? "codemaster-selected" : "player-selected"}>
                    <button onClick={(e) => this.toggleRole(e, 'player')} className="player">Player</button>
                    <button onClick={(e) => this.toggleRole(e, 'codemaster')} className="codemaster">Spymaster</button>
                    <button onClick={(e) => this.nextGame(e)} id="next-game-btn">Next game</button>
                </form>
            </div>
        );
    }
});
