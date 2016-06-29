window.App = React.createClass({
    getInitialState: function() {
        if (document.location.hash) {
            return {gameID: document.location.hash.slice(1)};
        }
        return {gameID: null};
    },

    gameSelected: function(game) {
        this.setState({gameID: game.id});
        document.location.hash = '#' + game.id;
    },

    render: function() {
        let pane;
        if (this.state.gameID) {
            pane = (<window.Game gameID={this.state.gameID} />)
        } else {
            pane = (<window.Lobby gameSelected={this.gameSelected} />)
        }

        return (
            <div id="application">
                <div id="topbar">
                    <h1>Codenames</h1>
                </div>
                {pane}
            </div>
        );
    }
});
