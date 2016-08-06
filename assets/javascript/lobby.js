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
                        <p className="intro">
                           This app allows you to play Codenames across multiple devices
                           with a shared board. To create a new game, enter a name and click
                           'New.'
                        </p>
                        <input type="text" id="game-name" placeholder="My game name" autoFocus
                            onChange={this.newGameTextChange} value={this.state.newGameName} />
                        <button onClick={this.handleNewGame}>New</button>
                    </form>
                    { this.state.games.length ? (<h3>Recent games</h3>) : null }
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
