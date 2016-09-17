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
                           This app allows you to play Codenames across multiple devices
                           with a shared board. To create a new game, click 'New.'
                        </p>
                        <input type="text" id="game-name" autoFocus
                            onChange={this.newGameTextChange} value={this.state.newGameName} />
                        <button onClick={this.handleNewGame}>New</button>
                    </form>
                </div>
            </div>
        );
    }
});
