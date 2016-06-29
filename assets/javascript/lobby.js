window.Lobby = React.createClass({
    propTypes: {
        gameSelected: React.PropTypes.func,
    },

    getInitialState: function() {
        return {
            newGameName: '',
            selectedGame: null,
            games: [],
        };
    },

    componentWillMount: function() {
      $.get('/games', (data) => { this.setState({games: data}); });
    },

    newGameTextChange: function(e) {
        this.setState({newGameName: e.target.value});
    },

    handleNewGame: function(e) {
        e.preventDefault();
        if (!this.state.newGameName) {
            return;
        }

        $.post('/new', JSON.stringify({name: this.state.newGameName}), this.joinGame);
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
                        <input type="text" id="game-name" placeholder="Drake & gamez" autoFocus
                            onChange={this.newGameTextChange} value={this.state.newGameName} />
                        <button onClick={this.handleNewGame}>New</button>
                    </form>
                    <ul>
                        { this.state.games.map((g) => (
                            <li key={g.id} onClick={() => this.joinGame(g)}>
                                {g.name}
                            </li>
                        )) }
                    </ul>
                </div>
            </div>
        );
    }
});
