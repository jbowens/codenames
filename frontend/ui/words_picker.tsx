import * as React from 'react';

export const OriginalWords =
  'AFRICA, AGENT, AIR, ALIEN, ALPS, AMAZON, AMBULANCE, AMERICA, ANGEL, ANTARCTICA, APPLE, ARM, ATLANTIS, AUSTRALIA, AZTEC, BACK, BALL, BAND, BANK, BAR, BARK, BAT, BATTERY, BEACH, BEAR, BEAT, BED, BEIJING, BELL, BELT, BERLIN, BERMUDA, BERRY, BILL, BLOCK, BOARD, BOLT, BOMB, BOND, BOOM, BOOT, BOTTLE, BOW, BOX, BRIDGE, BRUSH, BUCK, BUFFALO, BUG, BUGLE, BUTTON, CALF, CANADA, CAP, CAPITAL, CAR, CARD, CARROT, CASINO, CAST, CAT, CELL, CENTAUR, CENTER, CHAIR, CHANGE, CHARGE, CHECK, CHEST, CHICK, CHINA, CHOCOLATE, CHURCH, CIRCLE, CLIFF, CLOAK, CLUB, CODE, COLD, COMIC, COMPOUND, CONCERT, CONDUCTOR, CONTRACT, COOK, COPPER, COTTON, COURT, COVER, CRANE, CRASH, CRICKET, CROSS, CROWN, CYCLE, CZECH, DANCE, DATE, DAY, DEATH, DECK, DEGREE, DIAMOND, DICE, DINOSAUR, DISEASE, DOCTOR, DOG, DRAFT, DRAGON, DRESS, DRILL, DROP, DUCK, DWARF, EAGLE, EGYPT, EMBASSY, ENGINE, ENGLAND, EUROPE, EYE, FACE, FAIR, FALL, FAN, FENCE, FIELD, FIGHTER, FIGURE, FILE, FILM, FIRE, FISH, FLUTE, FLY, FOOT, FORCE, FOREST, FORK, FRANCE, GAME, GAS, GENIUS, GERMANY, GHOST, GIANT, GLASS, GLOVE, GOLD, GRACE, GRASS, GREECE, GREEN, GROUND, HAM, HAND, HAWK, HEAD, HEART, HELICOPTER, HIMALAYAS, HOLE, HOLLYWOOD, HONEY, HOOD, HOOK, HORN, HORSE, HORSESHOE, HOSPITAL, HOTEL, ICE, ICE CREAM, INDIA, IRON, IVORY, JACK, JAM, JET, JUPITER, KANGAROO, KETCHUP, KEY, KID, KING, KIWI, KNIFE, KNIGHT, LAB, LAP, LASER, LAWYER, LEAD, LEMON, LEPRECHAUN, LIFE, LIGHT, LIMOUSINE, LINE, LINK, LION, LITTER, LOCH NESS, LOCK, LOG, LONDON, LUCK, MAIL, MAMMOTH, MAPLE, MARBLE, MARCH, MASS, MATCH, MERCURY, MEXICO, MICROSCOPE, MILLIONAIRE, MINE, MINT, MISSILE, MODEL, MOLE, MOON, MOSCOW, MOUNT, MOUSE, MOUTH, MUG, NAIL, NEEDLE, NET, NEW YORK, NIGHT, NINJA, NOTE, NOVEL, NURSE, NUT, OCTOPUS, OIL, OLIVE, OLYMPUS, OPERA, ORANGE, ORGAN, PALM, PAN, PANTS, PAPER, PARACHUTE, PARK, PART, PASS, PASTE, PENGUIN, PHOENIX, PIANO, PIE, PILOT, PIN, PIPE, PIRATE, PISTOL, PIT, PITCH, PLANE, PLASTIC, PLATE, PLATYPUS, PLAY, PLOT, POINT, POISON, POLE, POLICE, POOL, PORT, POST, POUND, PRESS, PRINCESS, PUMPKIN, PUPIL, PYRAMID, QUEEN, RABBIT, RACKET, RAY, REVOLUTION, RING, ROBIN, ROBOT, ROCK, ROME, ROOT, ROSE, ROULETTE, ROUND, ROW, RULER, SATELLITE, SATURN, SCALE, SCHOOL, SCIENTIST, SCORPION, SCREEN, SCUBA DIVER, SEAL, SERVER, SHADOW, SHAKESPEARE, SHARK, SHIP, SHOE, SHOP, SHOT, SINK, SKYSCRAPER, SLIP, SLUG, SMUGGLER, SNOW, SNOWMAN, SOCK, SOLDIER, SOUL, SOUND, SPACE, SPELL, SPIDER, SPIKE, SPINE, SPOT, SPRING, SPY, SQUARE, STADIUM, STAFF, STAR, STATE, STICK, STOCK, STRAW, STREAM, STRIKE, STRING, SUB, SUIT, SUPERHERO, SWING, SWITCH, TABLE, TABLET, TAG, TAIL, TAP, TEACHER, TELESCOPE, TEMPLE, THEATER, THIEF, THUMB, TICK, TIE, TIME, TOKYO, TOOTH, TORCH, TOWER, TRACK, TRAIN, TRIANGLE, TRIP, TRUNK, TUBE, TURKEY, UNDERTAKER, UNICORN, VACUUM, VAN, VET, WAKE, WALL, WAR, WASHER, WASHINGTON, WATCH, WATER, WAVE, WEB, WELL, WHALE, WHIP, WIND, WITCH, WORM, YARD';

export class WordsPicker extends React.Component {
  constructor(props) {
    super(props);
    this.state = { default: true, expanded: false };
  }

  public toggle(e) {
    e.preventDefault();
    this.setState((state, props) => {
      return { expanded: !state.expanded };
    });
  }

  public onChange(e) {
    const text = e.target.value;
    this.props.onChange(text);
  }

  public render() {
    let symbol = this.state.expanded ? '▾' : '▸';
    let label = this.props.words == OriginalWords ? 'English' : 'Custom';

    let words = this.props.words
      .trim()
      .split(',')
      .map(w => w.trim())
      .filter(w => w.length > 0);

    let warning = null;
    if (words.length < 25) {
      warning = <div className="warning">must have 25+ words</div>;
    }

    let input = null;
    if (this.state.expanded) {
      input = (
        <div>
          <textarea
            value={this.props.words}
            onChange={this.onChange.bind(this)}
          />
        </div>
      );
    }

    return (
      <div>
        <div className="btn-configured-word-set" onClick={e => this.toggle(e)}>
          <div className="symbol">{symbol}</div>
          <div className="label">Words: {label}</div>
          {warning}
        </div>
        {input}
      </div>
    );
  }
}
