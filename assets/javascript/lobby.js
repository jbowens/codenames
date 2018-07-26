window.Lobby = React.createClass({
    propTypes: {
        gameSelected:   React.PropTypes.func,
        defaultGameID: React.PropTypes.string,
    },

    getInitialState: function() {
        return {
            newGameName: this.props.defaultGameID,
            selectedGame: null,
        };
    },

    newGameTextChange: function(e) {
        this.setState({newGameName: e.target.value});
    },

    handleNewGame: function(e) {
        e.preventDefault();
        if (!this.state.newGameName) {
            return;
        }

        $.post('/game/'+this.state.newGameName, this.joinGame);
        this.setState({newGameName: ''});
    },

    joinGame: function(g) {
        this.setState({selectedGame: g});
        if (this.props.gameSelected) {
            this.props.gameSelected(g);
        }
    },

    render: function() {
        return (
            <div id="lobby">
                <div id="available-games">
                    <form id="new-game">
                        <p className="intro">
                           Play Codenames online across multiple devices on a shared board.
                           To create a new game or join an existing
                           game, enter a game identifier and click 'GO'.
                        </p>
                        <input type="text" id="game-name" autoFocus
                            onChange={this.newGameTextChange} value={this.state.newGameName} />
                        <button onClick={this.handleNewGame}>Go</button>
                    </form>
                </div>
            </div>
        );
    }
});
