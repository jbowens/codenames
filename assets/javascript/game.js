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
          this.setState({game: data});
          setTimeout(this.refresh, 3000);
      });
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

    endTurn: function() {
        $.post('/end-turn', JSON.stringify({game_id: this.state.game.id}),
              (g) => { this.setState({game: g}); });
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

        return (
            <div id="game-view" className={this.state.codemaster ? "codemaster" : "player"}>
                <div id="share">
                  To connect another device, open this URL on the other device's browser: <a className="url" href={window.location.href}>{window.location.href}</a>. All devices on this URL will share the same board.
                </div>
                <div id="status-line" className={statusClass}>
                    <div id="status" className="status-text">{status}</div>
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
                </form>
            </div>
        );
    }
});
