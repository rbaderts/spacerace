
//#import * as PIXI from 'pixi.js'

//var PIXI = require('pixi.js');

import * as PIXI from 'pixi.js';
const loader = new PIXI.Loader();
import {Howl, Howler} from 'howler';
import {divide, multiply, evaluate, pow, abs, atan2, sin, cos, round, sqrt, add, subtract} from 'mathjs'

var flatbuffers = require('flatbuffers').flatbuffers;
var gamestate = require('./gamestate_generated.js').messages;
var playercommands = require('./playercommands_generated.js').messages;

//import * as PIXI from 'pixi.js'

var sprintf = require('sprintf-js').sprintf;

//var particles = require('pixi.particles.js');
//var SpriteUtilities = require("./SpriteUtilities.js");
//var pb = require('./messages_pb.js');


//import { ParticleContainer, loader } from 'pixi.js';
//import { Emitter } from 'pixi-particles';


const FORCE_COEFFICIENT = 1000;
const DISTANCE_COEFFICIENT = 100;

var GRAVITY = multiply(6.673, pow(10, -11));
var TWO_PI = evaluate("2*PI");

const EXPLOSION_DURATION = 1000;
const FPS = 60;
const Viewport = require('pixi-viewport').Viewport;

const SettingsPanel = require("settingspanel");
const Debug = require("yy-debug");
const WorldMap = require("./world_map.js");

const SPRITE_KIND = 0xff0000;
const SHIP_STATE = 0x0000ff;
const PRIZE_TYPE = 0x00ff00;
const PRIZE_VALUE = 0x0000ff;
const PLAY_SOUNDS = true;
const PLAY_BACKGROUND_MUSIC = true;

const JETS_ON = 0x000001;
const SHIELDS_ACTIVE = 0x000002;
const PHANTOM_MODE = 0x000004;
const CLOAK_MODE = 0x000008;
const TRACTOR_ACTIVE = 0x000010;

var pretty = require("js-object-pretty-print").pretty;

var TheGame = null;
var GameOver = false;
var FreezeDrawing = false;
var WS = null;
var OurShip = null;
var OurShipId = null;
//var PhysicsUpdateBuffer = new UpdateBuffer();
var PhysicsUpdateBuffer = new RingBuffer(10);
var ShipUpdateBuffer = new RingBuffer(10);

var PlayerUpdateQueue = new Array();
var PlaySoundQueue = new Array();
var ShakeQueue = new Array();
var OutQueue = new Array();

var theBackground = null;
var theWorldMap = null;
var WorldMapApplication = null;

var GameId = document.getElementById("GameID").value;
var PlayerId = document.getElementById("PlayerID").value;
var ActionId = 0;
var LatestUnresolvedActionId = 0;

var READY = false;

var ShieldTokens = 0;
var HyperspaceTokens = 0;
var BoosterTokens = 0;
var CloakTokens = 0;
var BlackholeMass = 1000000000000;
var BulletSpeed = 60;

const HEIGHT = 5000;
const WIDTH = 5000;

var VIEW_HEIGHT = 910;
var VIEW_WIDTH = 1050;

var OVERVIEW_HEIGHT = 280;
var OVERVIEW_WIDTH = 280;

// Scrolling Background

function Background(width, height, x, y) {
    this.texture = PIXI.Texture.from("/static/img/BG2.gif");
    PIXI.TilingSprite.call(this, this.texture, width, height);

    this.position.x = 0;
    this.position.y = 0;
    this.tilePosition.x = -x + VIEW_WIDTH / 2;
    this.tilePosition.y = -y + VIEW_HEIGHT / 2;
    this.viewportX = -x + VIEW_WIDTH / 2;
    this.viewportY = -y + VIEW_HEIGHT / 2;
}

Background.prototype = Object.create(PIXI.TilingSprite.prototype);

Background.prototype.scrollTo = function (x, y) {
    this.viewportX = -x - VIEW_WIDTH / 2;
    this.viewportY = -y - VIEW_HEIGHT / 2;
    this.tilePosition.x = -x - VIEW_WIDTH / 2;
    this.tilePosition.y = -y - VIEW_HEIGHT / 2;
    this.position.x = 0;
    this.position.y = 0;
};

Background.prototype.resize = function (x, y) {};

//var overviewRenderer = null;
//var overviewStage = null;
var AppViewport = null;
var WorldMapViewport = null;
//var graphics = null;
var Application = null;

function setupWorldMap(Images) {

    let overviewDiv = document.getElementById("overview_div");
    let width = overviewDiv.clientWidth;
    let height = overviewDiv.clientHeight;

    WorldMapApplication = new PIXI.Application(width, height, { forceCanvas: true }, true, false);

    overviewDiv.appendChild(WorldMapApplication.view);

    WorldMapViewport = new Viewport();
    WorldMapViewport.resize(OVERVIEW_WIDTH, OVERVIEW_HEIGHT, WIDTH, HEIGHT);


    OVERVIEW_WIDTH = width;
    OVERVIEW_HEIGHT = height;
    theWorldMap = new WorldMap(WorldMapViewport, OVERVIEW_HEIGHT, OVERVIEW_WIDTH, HEIGHT, WIDTH); //    WorldMapApplication.stage.addChild(theWorldMap);

    console.log(sprintf("worldmap: %d %d %d %d\n", OVERVIEW_HEIGHT, OVERVIEW_WIDTH, HEIGHT, WIDTH));

    // graphics = new PIXI.Graphics();
    // graphics.x = 0;
    // graphics.y = 0;
    //theWorldMap.addChild(graphics)
    //WorldMapApplication.stage.addChild(graphics);
    WorldMapApplication.stage.addChild(WorldMapViewport);
    //    WorldMapStage = WorldMapApplication.stage;


    //setup();
}

function setupGameArea() {

    let gameDiv = document.getElementById("gamearea_div");

    VIEW_WIDTH = gameDiv.clientWidth;
    VIEW_HEIGHT = gameDiv.clientHeight;
    Application = new PIXI.Application(VIEW_WIDTH, VIEW_HEIGHT, { forceCanvas: true }, true, false);
    Application.renderer.clearBeforeRender = true;

    gameDiv.appendChild(Application.view);
    theBackground = new Background(WIDTH, HEIGHT, WIDTH / 2, HEIGHT / 2);

    AppViewport = new Viewport();
    AppViewport.resize(VIEW_WIDTH, VIEW_HEIGHT, WIDTH, HEIGHT);

    Application.stage.addChild(AppViewport);
    AppViewport.addChild(theBackground);
}

// DebugPanels

var shipSpritePos = null;
var dropped = null;
var viewportCenter = null;

function setupDebugPanels() {
    Debug.init();

    dropped = Debug.add("Dropped", { text: "Dropped: 0", side: "leftBottom" });
    shipSpritePos = Debug.add("ShipSpritePos", {
        text: "Sprite: 0, 0",
        side: "leftBottom"
    });
    viewportCenter = Debug.add("ViewportCenter", {
        text: "Viewport: 0, 0",
        side: "leftBottom"
    });

    // SettingPanel

    const panel = new SettingsPanel({ background: "rgb(50,50,50)" });

    // create a button that changes its own color on callback and increments a counter on the button
    panel.input("Blackhole Mass", function (value) {
        var msg = new ClientMessageData(playercommands.CommandType.SetBlackholeMass, value);
        OutQueue.push(msg);
    }, { original: BlackholeMass, color: "red" });

    panel.input("Bullet Speed", function (value) {
        var msg = new ClientMessageData(playercommands.CommandType.SetBulletSpeed, value);
        OutQueue.push(msg);
    }, { original: BlackholeMass, color: "red" });
}

var rotationDisplay = document.getElementById("rotation-display");

var HealthProgress = new ProgressBar("lifeenergy", 100);
var ShieldsProgress = new ProgressBar("shields_progress", 10);
var BoostProgress = new ProgressBar("boost_progress", 10);
var HyperspaceProgress = new ProgressBar("hyperspace_progress", 10);

function updateGauges() {

    if (OurShip != null) {
        ShieldsProgress.update(ShieldTokens);
        BoostProgress.update(BoosterTokens);
        HyperspaceProgress.update(HyperspaceTokens);

        if (OurShip.healthPoints != null) {
            HealthProgress.update(OurShip.healthPoints);
        }
        var r = OurShip.sprite.rotation;
        rotationDisplay.innerHTML = r.toFixed(3).toString();
    }
}

/*
var ParticleEmitter = null
function setupParticles() {

    useParticleContainer = true
    var emitterContainer;
    if(useParticleContainer)
    {
        emitterContainer = new PIXI.ParticleContainer();
        emitterContainer.setProperties({
            scale: true,
            position: true,
            rotation: true,
            uvs: true,
            alpha: true
        });
    }
    else
        emitterContainer = new PIXI.Container();

    stage.addChild(emitterContainer);

    config = {
            alpha: {
                start: 0.8,
                end: 0.1
            },
            scale: {
                start: 1,
                end: 0.3
            },
            color: {
                start: "fb1010",
                end: "f5b830"
            },
            speed: {
                start: 200,
                end: 100
            },
            startRotation: {
                min: 0,
                max: 360
            },
            rotationSpeed: {
                min: 0,
                max: 0
            },
            lifetime: {
                min: 0.5,
                max: 0.5
            },
            frequency: 0.008,
            emitterLifetime: 0.31,
            maxParticles: 1000,
            pos: {
                x: 0,
                y: 0
            },
            addAtBack: false,
            spawnType: "circle,"
            spawnCircle: {
                x: 0,
                y: 0,
                r: 10
            }
        };

    ParticleEmitter = new PIXI.particles.Emitter(
        emitterContainer,
        art,
        config
    );
    if(type == "path")
        emitter.particleConstructor = PIXI.particles.PathParticle;
    else if(type == "anim")
        emitter.particleConstructor = PIXI.particles.AnimatedParticle;

    // Center on the stage
    emitter.updateOwnerPos(window.innerWidth / 2, window.innerHeight / 2);

}
*/

// Setup the new Howl.
const thrustSound = new Howl({
    src: ['/static/snd/thrust2.wav']
});
const explosionSound = new Howl({
    src: ['/static/snd/explosion1.wav']
});
const laserSound = new Howl({
    src: ['/static/snd/laser.wav']
});
const clickSound = new Howl({
    src: ['/static/snd/click.wav']
});
const shieldSound = new Howl({
    src: ['/static/snd/shield.wav']
});
const boingSound = new Howl({
    src: ['/static/snd/boing.wav']
});
const backgroundSound = new Howl({
    src: ['/static/snd/BackgroundMusic.wav']
});

/*
function round(num) {
    //return math.round(num * 1000) / 1000;
    return divide(round(multiply(num, 1000)), 1000);
}
*/

function startBackgroundMusic() {
    backgroundSound.play();
}

function dynamicCall(func) {
    this[func].apply(this, Array.prototype.slice.call(arguments, 1));
}

var Images = null;
loader.add("/static/img/sheet.json").load(setup);

function setup() {
    Images = loader.resources["/static/img/sheet.json"].textures;
    setupGameArea();
    setupWorldMap(Images);
    setupDebugPanels();
    Application.renderer.render(Application.stage);
    WorldMapApplication.renderer.render(WorldMapApplication.stage);
    startGame();
    READY = true;
    requestAnimationFrame(gameLoop);
}

var frameOffset = -1;
var counter1 = 0;

function bulletSpeedChange(element) {
    var speed = element.val();
    var msg = new ClientMessageData(playercommands.CommandType.SetBulletSpeed, 0.1);
    OutQueue.push(msg);
}

function Now() {
    return performance.now();
}

var T = 0;
var LastFrame = 0;
var Frame = 0;
var StartTime = Now();
var droppedUpdates = 1;

var dt = evaluate(sprintf("1/%f", FPS));
var accumulator = 0.0;

function gameLoop(timestamp) {

    var now = Now();

    if (Frame % 30 == 0) {
        updateGauges();
    }

    if (T === 0) {
        T = timestamp;
        LastFrame = timestamp;
    }

    var frameTimeMillis = timestamp - LastFrame; // get the delta time since last frame
    LastFrame = timestamp;

    var delta = divide(frameTimeMillis, 1000);

    if (FreezeDrawing == false) {

        update(delta, Frame);
        draw();
    }

    T += frameTimeMillis;
    Frame += 1;

    sendUpdates(delta);

    updateDebugs();

    requestAnimationFrame(gameLoop);
}

var DebugCount = 0;

function updateDebugs() {
    if (DebugCount != 20) {
        DebugCount += 1;
        return;
    }
    DebugCount = 0;

    if (StartTime > 0 && droppedUpdates > 0) {
        var elapsedMillis = Now() - StartTime;

        var exp = sprintf("%d/(%f/1000)", droppedUpdates, elapsedMillis);
        var rs = evaluate(exp);

        //var rs = math.divide(droppedUpdates, math.divide(elapsedMillis, 1000));

        var msg = "Avg Updates Dropped per second: " + rs.toFixed(2).toString();
        Debug.one(msg, { panel: dropped, color: "red" });
    }
    if (OurShip != null) {
        var pos = "Sprite: (" + OurShip.sprite.x.toFixed(2).toString() + ", " + OurShip.sprite.y.toFixed(2).toString() + ")";
        Debug.one(pos, { panel: shipSpritePos, color: "red" });
    }
    if (AppViewport != null) {
        var pos = "Viewport Center: (" + AppViewport.center.x.toFixed(2) + ", " + AppViewport.center.y.toFixed(2) + ")";
        Debug.one(pos, { panel: viewportCenter, color: "red" });
    }
}

function sendUpdates(delta) {

    if (OutQueue.length > 0) {
        console.log("Sending " + OutQueue.length + " updates");

        var fbb = new flatbuffers.Builder(64);
        let Offsets = new Array();

        while (OutQueue.length > 0) {
            let item = OutQueue.pop();
            playercommands.PlayerCommandMessage.startPlayerCommandMessage(fbb);
            playercommands.PlayerCommandMessage.addCmd(fbb, item.cmd);
            playercommands.PlayerCommandMessage.addValue(fbb, item.value);
            playercommands.PlayerCommandMessage.addActionId(fbb, item.actionId);
            let offset = playercommands.PlayerCommandMessage.endPlayerCommandMessage(fbb);
            Offsets.push(offset);
        }

        console.log("adding len messages: " + Offsets.length);
        let msgsOffset = playercommands.ClientMessage.createMessagesVector(fbb, Offsets);

        playercommands.ClientMessage.startClientMessage(fbb);
        playercommands.ClientMessage.addMessages(fbb, msgsOffset);
        let finalOffset = playercommands.ClientMessage.endClientMessage(fbb);
        fbb.finish(finalOffset);

        //playercommands.ClientMessage.finishClientMessageBuffer(fbb, finalOffset)

        let buf = fbb.dataBuffer();
        let bytes = buf.bytes().subarray(buf.position(), buf.position() + buf.bytes().length);
        send(bytes);
        OutQueue = new Array();
    }
}

function send(data) {
    if (WS != null) {
        WS.send(data);
    }
}

function update(delta, frame) {

    if (counter1 == 5) {
        counter1 = 0;
        if (leftArrow.isDown) {
            OurShip.rotate(-0.1);
            //var msg = BuildClientMessage(gamestate.CommandType.Rotate, -0.1);
            var msg = new ClientMessageData(playercommands.CommandType.Rotate, -0.1);
            OutQueue.push(msg);
        } else if (rightArrow.isDown) {
            OurShip.rotate(0.1);
            var msg = new ClientMessageData(playercommands.CommandType.Rotate, 0.1);
            OutQueue.push(msg);
        }
    }

    counter1 += 1;

    var playerUpdates = PlayerUpdateQueue.length;
    if (playerUpdates > 0) {
        for (var i = 0; i < playerUpdates; i++) {
            console.log("process Player Update");
            processPlayerUpdate(PlayerUpdateQueue[i]);
        }
        updateGauges();
        PlayerUpdateQueue = new Array();
    }

    console.log("updating sprites for frame: " + frame + ", elapsed time = " + delta);
    updateSprites(frame);

    var highestFrame = 0;

    // Calculate local physics

    //TheGame.computeGravityPullOnPlayerShip();

    if (OurShip != null) {
        OurShip.move(delta, frame);
        centerViewOnShip(OurShip);
    }

    //    for (var i = 0; i < msgs; i++) {
    //updateMessage = PhysicsUpdateQueue[i];
    //        updateMessage = PhysicsUpdateQueue[i];
    //       if (updateMessage.getFrame() > highestFrame) {
    //           highestFrame = updateMessage.getFrame()
    //       }
    //   }

    ///    updateMessage = PhysicsUpdateBuffer.pull(0);

    AppViewport.update();

    // Play Sounds
    if (PLAY_SOUNDS == true) {
        var sounds = PlaySoundQueue.length;
        for (i = 0; i < sounds; i++) {
            handlePlaySound(PlaySoundQueue[i]);
        }
        PlaySoundQueue = new Array();
    }

    // Shake sprites
    var shakeCount = ShakeQueue.length;
    for (i = 0; i < shakeCount; i++) {
        handleShake(ShakeQueue[i]);
    }
    ShakeQueue = new Array();
}

function centerViewOnShip(OurShip) {

    if (AppViewport != null) {

        var x = OurShip.sprite.position.x;
        var y = OurShip.sprite.position.y;

        AppViewport.moveCenter(x, y);
    }
}

function processPlayerUpdate(playerUpdate) {

    console.log("playerUpdate = " + playerUpdate);

    let len = playerUpdate.inventoryLength();

    //var inventories = player.getInventoryList();

    for (var i = 0; i < len; i++) {

        //        var inventory = inventories[i]
        var inventory = playerUpdate.inventory(i, null);
        if (inventory.resourceType() == gamestate.PlayerResourceType.Shield) {
            ShieldTokens = inventory.value();
        } else if (inventory.resourceType() == gamestate.PlayerResourceType.Booster) {
            BoosterTokens = inventory.value();
        } else if (inventory.resourceType() == gamestate.PlayerResourceType.Hyperdrive) {
            HyperspaceTokens = inventory.value();
        } else if (inventory.resourceType() == gamestate.PlayerResourceType.Cloak) {
            CloakTokens = inventory.value();
        }
    }
}

function scrollBackground() {}

function draw() {
    Application.renderer.render(Application.stage);
    WorldMapApplication.renderer.render(WorldMapApplication.stage);
}

function updateSprites(frame) {

    let updateMessage = PhysicsUpdateBuffer.pull(0);
    if (!updateMessage) {
        return;
    }

    let serverFrame = updateMessage.frame();
    let serverFrameTime = updateMessage.frameTime();
    console.log("Server Frame: " + serverFrame + ", serverFrameTime: " + serverFrameTime);

    var SpriteList = new Map();
    var actionId = updateMessage.actionId();

    if (actionId >= LatestUnresolvedActionId) {
        LatestUnresolvedActionId = 0;
    }

    for (let [key, value] of TheGame.all) {
        SpriteList.set(key, value);
    }

    var numSprites = updateMessage.spritesLength();


    //for (var e in sprites.entries()) {
    for (let i = 0; i < numSprites; ++i) {
        let entry = updateMessage.sprites(i, null);
        TheGame.handleSprite(entry, actionId);
        let id = entry.id();
        SpriteList.delete(id);
    }

    for (let [key, value] of SpriteList) {
        //for (var s in SpriteList) {
        var gameobj = value;

        WorldMapViewport.removeChild(gameobj.wm_sprite);
        AppViewport.removeChild(gameobj.sprite);
        TheGame.all.delete(key);
    }
}

jQuery("#newgame").on("click", function () {
    jQuery.get("/newgame", function (data) {
        startGame();
    });
});

function startGame() {
    runGame();
}

function runGame() {

    TheGame = new Game();
    TheGame.init();

    WS = new WebSocket("ws://" + window.location.host + "/updates/" + GameId);
    WS.binaryType = "arraybuffer";

    WS.onopen = function () {
        console.log("Server connection enstablished");
    };

    WS.onclose = function () {
        console.log("Server close");
        WS = null;
        var page = "/lobby";
        window.location.assign(page);
    };

    WS.onerror = function (error) {
        console.log("Server error: " + error);
        WS = null;
        var page = "/lobby";
        window.location.assign(page);
    };

    WS.onmessage = function (msg) {

        if (TheGame == null) {
            console.log("no Game");
            return;
        }
        handleServerCommand(msg);
    };
}

function handleServerCommand(updatemsg) {

    var data = new Uint8Array(updatemsg.data);

    var bb = new flatbuffers.ByteBuffer(data);

    var update = gamestate.Update.getRootAsUpdate(bb, null);


    if (update.messageType() == gamestate.UpdateMessage.PhysicsUpdate) {
        let msg = update.message(new gamestate.PhysicsUpdate());
        PhysicsUpdateBuffer.push(msg);
    } else if (update.messageType() == gamestate.UpdateMessage.PlayerUpdate) {
        let msg = update.message(new gamestate.PlayerUpdate());
        PlayerUpdateQueue.push(msg);
    } else if (update.messageType() == gamestate.UpdateMessage.InitializePlayer) {
        let msg = update.message(new gamestate.InitializePlayer());
        handlePlayerInitialize(msg);
    } else if (update.messageType() == gamestate.UpdateMessage.PlaySound) {
        let msg = update.message(new gamestate.PlaySound());
        if (PLAY_SOUNDS == true) {
            PlaySoundQueue.push(msg);
        }
    } else if (update.messageType() == gamestate.UpdateMessage.Draw) {
        let msg = update.message(new gamestate.Draw());
        handleDrawMessage(msg);
    } else if (update.messageType() == gamestate.UpdateMessage.Shake) {
        let msg = update.message(new gamestate.Shake());
        ShakeQueue.push(msg);
    } else if (update.messageType() == gamestate.UpdateMessage.PlayerDead) {
        let msg = update.message(new gamestate.PlayerDead());
        handlePlayerDead(msg);
    }
    //  console.log("Received server message type: " + message.typ);
    //  console.log("message = " + JSON.stringify(message));
    //  console.log("msg.data = " + JSON.stringify(msg.data));
    // }
}

function handleShake(shake) {
    var spriteId = shake.spriteId();
    var magnitude = shake.magnitude() / 100;

    var sp = TheGame.all.get(spriteId);

    if (sp) {
        //SpriteUtils.shake(sp.sprite, magnitude, false);
    } else {
        console.log("Null gameobject for spriteId: " + spriteId);
    }
}

function handlePlayerInitialize(p) {
    console.log("handlePlayerInitialize: playerId: " + p.playerId() + ", shipId: " + p.shipId());
    PlayerId = p.playerId();
    OurShipId = p.shipId();
}

function playBackgroundMusic() {}

function handlePlaySound(sound) {

    console.log("handlePlaySound");
    if (PLAY_SOUNDS == false) {
        return;
    }
    if (sound.soundType() == gamestate.SoundType.ExplosionSound) {
        explosionSound.volume = sound.volume();
        explosionSound.play();
    } else if (sound.soundType() == gamestate.SoundType.BoingSound) {
        boingSound.volume = sound.volume();
        boingSound.play();
    } else if (sound.soundType() == gamestate.SoundType.BloopSound) {
        bloopSound.volume = sound.volume();
        bloopSound.play();
    }
}

function handlePlayerDead(dead) {

    if (dead.playerId() == PlayerId) {

        OurShip.sprite.position.x = WIDTH / 2;
        OurShip.sprite.position.y = HEIGHT / 2;
        OurShip.sprite.rotation = 0;
        Application.renderer.render(Application.stage);
    }

}

function eraseTractor(ship) {
    graphics.clear();
}
function handleDrawMessage(draw) {
    var graphics = new PIXI.Graphics();
    var cmds = draw.getCmdsList();
    var len = cmds.length;

    for (var i = 0; i < len; i++) {
        //console.log("cmd = " + cmds[i]);
        eval("graphics." + cmds[i]);
    }
    Application.renderer.render(Application.stage);
}

function hasState(stateList, state) {
    var len = stateList.length;
    for (var i = 0; i < len; i++) {
        if (stateList[i] == state) {
            return true;
        }
    }
    return false;
}

function hasProperty(propertyList, prop) {
    var len = propertyList.length;
    for (var i = 0; i < len; i++) {
        if (propertyList[i] == prop) {
            return true;
        }
    }
    return false;
}

var leftArrow = keyboard(37);
var upArrow = keyboard(38);
var rightArrow = keyboard(39);
var downArrow = keyboard(40);
var space = keyboard(32);
var esc = keyboard(27);
var tab = keyboard(9);
var control = keyboard(17);
var s = keyboard(83);
var w = keyboard(87);
var c = keyboard(67);
var t = keyboard(84);

function Game() {

    this.state = this.play;
    this.all = new Map();
    this.shipSpriteBuffer = [];
    this.jetsCounter = 0;

    this.OurShipSpriteUpdates = new RingBuffer(10);

    this.init = function () {

        leftArrow.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.Rotate, -0.1);

            OurShip.rotate(-0.1);
            OutQueue.push(msg);
        };

        rightArrow.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.Rotate, 0.1);
            OurShip.rotate(0.1);
            OutQueue.push(msg);
        };

        upArrow.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.Thrust, 100);
            var actionId = msg.getActionId();
            var force = new Force(OurShip.sprite.rotation, 100, actionId, 100);
            console.log("Thrust: direction = " + OurShip.sprite.rotation);

            OurShip.addForce(force);

            OutQueue.push(msg);
            if (PLAY_SOUNDS == true) {
                thrustSound.play();
            }
        };

        tab.press = function () {
            if (BoosterTokens >= 1) {
                var msg = new ClientMessageData(playercommands.CommandType.Boost, 200);
                var actionId = msg.getActionId();
                var force = new Force(OurShip.sprite.rotation, 200, actionId, 200);
                OurShip.addForce(force);

                OutQueue.push(msg);
                //double x = 3;
                //double y = 4;
                if (PLAY_SOUNDS == true) {
                    thrustSound.volume(0.7);
                    thrustSound.play();
                    thrustSound.volume(0.5);
                }
            }
        };

        s.press = function () {
            console.log("s pressed, shieldTokens = " + ShieldTokens);
            if (ShieldTokens > 1) {
                var msg = new ClientMessageData(playercommands.CommandType.ShieldOn, 5);
                OutQueue.push(msg);
                if (PLAY_SOUNDS == true) {
                    shieldSound.play();
                }
            } else {
                console.log("No Shields");
            }
        };

        s.release = function () {
            var msg = new ClientMessageData(playercommands.CommandType.ShieldOff, 5);
            OutQueue.push(msg);
            if (PLAY_SOUNDS == true) {
                shieldSound.stop();
            }
        };

        w.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.Hyperspace, 5);
            OutQueue.push(msg);
        };

        c.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.CloakShip, 5);
            OutQueue.push(msg);
        };

        t.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.TractorOn, 5);
            OutQueue.push(msg);
        };

        t.release = function () {
            var msg = new ClientMessageData(playercommands.CommandType.TractorOff, 5);
            OutQueue.push(msg);
        };

        downArrow.press = function () {};

        space.press = function () {
            var msg = new ClientMessageData(playercommands.CommandType.Fire, 700);
            OutQueue.push(msg);
            if (PLAY_SOUNDS == true) {
                laserSound.play();
            }
        };
    };

    this.handleSprite = function (sprite, actionId) {
        let id = sprite.id();
        if (!this.all.has(id)) {
            this.newSprite(sprite, actionId);
        } else {
            this.updateSprite(sprite, actionId);
        }
    };

    this.newSprite = function (sprite, actionId) {

        var img = null;
        var kind = sprite.typ() & SPRITE_KIND;
        console.log("newSprite: id = " + sprite.id() + ", kind = " + kind);

        var wm_img = null;
        if (kind === gamestate.SpriteKind.LargeAsteroid) {
            img = "largeroid_2.gif";
            wm_img = "WM_LargeRoid_2.png";
        } else if (kind === gamestate.SpriteKind.SmallAsteroid) {
            img = "smallroid_2.gif";
            wm_img = "WM_SmallRoid_2.png";
        } else if (kind === gamestate.SpriteKind.SpaceStation) {
            img = "spacestation.gif";
            wm_img = "WM_Spacestation.png";
        } else if (kind === gamestate.SpriteKind.Ship) {
            img = "SS.gif";
            wm_img = "WM_PlayerShip_2.png";
        } else if (kind === gamestate.SpriteKind.AiShip) {
            img = "SS.gif";
            wm_img = "WM_AIShip_2.png";
        } else if (kind === gamestate.SpriteKind.Bullet) {
            img = "bullet.gif";
            wm_img = "WM_Bullet_1.gif";
        } else if (kind === gamestate.SpriteKind.Blackhole) {
            img = "Blackhole.png";
            wm_img = "WM_Blackhole_3.png";
        } else if (kind === gamestate.SpriteKind.Star) {
            wm_img = "WM_Star_1.gif";
            img = "star.gif";
        } else if (kind === gamestate.SpriteKind.Prize) {
            img = "prize.gif";
            wm_img = "WM_Prize.gif";
        } else if (kind === gamestate.SpriteKind.Planet) {
            img = "planet.gif";
            wm_img = "WM_Planet_1.gif";
        } else if (kind === gamestate.SpriteKind.EndToken) {
            img = "endtoken.gif";
        } else {
            console.log("Unknown sprite kind: " + kind);
        }

        var gameobject = new GameObject(sprite, {
            typ: sprite.typ(),
            id: sprite.id(),
            img: img,
            wm_img: wm_img,
            height: sprite.height(),
            width: sprite.width(),
            xPos: sprite.x(),
            yPos: sprite.y(),
            mass: sprite.mass(),
            maxAge: 1000,
            playerName: sprite.playerName(),
            playerId: sprite.playerId(),
            healthPoints: sprite.healthpoints()
        });

        if (kind === gamestate.SpriteKind.Ship && sprite.playerId() === PlayerId) {
            console.log("OurShip = " + gameobject.id);
            OurShip = gameobject;
            centerViewOnShip(OurShip);
        }

        this.all.set(sprite.id(), gameobject);
    };

    this.updateSprite = function (sprite, actionId) {

        var img = null;
        var gameobject = this.all.get(sprite.id());

        if (gameobject == null) {
            console.log("gameobject is null for id: !" + sprite.id())
        }

        var ignoreShipUpdates = false;
        var ps = gameobject.sprite;

        let workingSprite = sprite;
        var shipState = workingSprite.typ() & SHIP_STATE;

        gameobject.healthPoints = workingSprite.healthpoints();

        if (gameobject != OurShip) {
            gameobject.setPosition(workingSprite.x(), workingSprite.y());
            gameobject.rotateTo(workingSprite.rotation());
            gameobject.vx = workingSprite.vx();
            gameobject.vy = workingSprite.vy();
        } else {

            var deltaX = round(ps.position.x - workingSprite.x());
            var deltaY = round(ps.position.y - workingSprite.y());

            console.log("ship sprite: location disagreement: (" + deltaX + "," + deltaY + ")");

            if (LatestUnresolvedActionId == 0) {
                gameobject.vx = workingSprite.vx();
                gameobject.vy = workingSprite.vy();
            }

            var dis = sqrt(pow(deltaX, 2) + pow(deltaY, 2));
            dis = round(dis);
            if (dis > 3) {
                gameobject.setPosition(workingSprite.x(), workingSprite.y());
                console.log("yanking sprite " + workingSprite.id() + " into position");
            } else if (dis > 0.0) {
                gameobject.setPosition(subtract(ps.position.x, multiply(deltaX, 0.2)), subtract(ps.position.y, multiply(deltaY, 0.2)));
            }

            var deltaRotation = abs(ps.rotation - workingSprite.rotation());

            if (deltaRotation > 1) {
                gameobject.rotateTo(workingSprite.rotation());
            } else if (deltaRotation > 0.1) {
                if (workingSprite.rotation() > gameobject.sprite.rotation) {
                    gameobject.rotateTo(gameobject.sprite.rotation + multiply(deltaRotation, 0.2));
                } else {
                    gameobject.rotateTo(gameobject.sprite.rotation - multiply(deltaRotation, 0.2));
                }
            }
        }

        if (gameobject.getKind() == gamestate.SpriteKind.Ship || gameobject.getKind() == gamestate.SpriteKind.AiShip) {

            gameobject.sprite.visible = true;

            if (gameobject != OurShip) {
                if (gameobject.isCloaked()) {
                    gameobject.sprite.visible = false;
                }
            } else {
                var shield = shipState & SHIELDS_ACTIVE;
                var jets = shipState & JETS_ON;
                var phantom = shipState & PHANTOM_MODE;
                var tractor = shipState & TRACTOR_ACTIVE;

                if (this.phantom != phantom) {
                    gameobject.setPhantom(phantom);
                }

                if (shield) {
                    gameobject.setShieldOn();
                } else {
                    gameobject.setShieldOff();
                }

                if (this.jetsCounter >= 5) {
                    this.jetsCounter = 0;
                }

                if (jets != 0 || this.jetsCounter > 0) {
                    img = "SSJ.gif";
                    this.jetsCounter += 1;
                } else {
                    if (this.jetsCounter == 0) {
                        img = "SS.gif";
                    }
                }

                if (img != gameobject.img) {
                    gameobject.img = img;
                    if (Images != null) {
                        ps.texture = Images[img];
                    }
                }

                if (gameobject.isCloaked()) {
                    ps.alpha = 0.2;
                } else {
                    ps.alpha = 1.0;
                }

                ps.height = ps.texture.height;
                ps.width = ps.texture.width;
            }
        } else if (gameobject.getKind() == gamestate.SpriteKind.SpaceStation) {
            img = "spacestation.gif";
            if (workingSprite.healthpoints() < 40) {
                img = "spacestation_3.gif";
            } else if (workingSprite.healthpoints() < 100) {
                img = "spacestation_2.gif";
            }
            if (img != gameobject.img) {
                gameobject.img = img;
                if (Images != null) {
                    ps.texture = Images[img];
                }
            }
        }
    };

    this.computeGravityPullOnPlayerShip = function () {
        for (let [key, value] of TheGame.all) {
            if (value != OurShip) {
                value.pullOn(OurShip);
            }
        }
    };
};

this.play = function () {};

this.keyboard = function (keyCode) {

    var key = {};
    key.code = keyCode;
    key.isDown = false;
    key.isUp = true;
    key.press = undefined;
    key.release = undefined;

    key.downHandler = function (event) {
        if (event.keyCode === key.code) {
            if (key.isUp && key.press) {
                key.press();
            } else if (event.repeat == true) {
                key.press();
            }
            key.isDown = true;
            key.isUp = false;
        }
        event.preventDefault();
    };

    key.upHandler = function (event) {
        if (event.keyCode === key.code) {
            if (key.isDown && key.release) key.release();
            key.isDown = false;
            key.isUp = true;
        }
        event.preventDefault();
    };

    //Attach event listeners
    window.addEventListener("keydown", key.downHandler.bind(key), false);
    window.addEventListener("keyup", key.upHandler.bind(key), false);
    return key;
};

var FRAME_RATE_NANOS = 16670000;

function Force(d, m, actionId, duration) {
    this.direction = d;
    this.magnitude = m;
    this.actionId = actionId;
    this.duration = duration;
    this.start = Now();

    this.getMagnitude = function () {
        return this.magnitude;
    };

    this.getDirection = function () {
        return this.direction;
    };

    this.getActionId = function () {
        return this.actionId;
    };

    this.getDuration = function () {
        return this.duration;
    };

    this.getStart = function () {
        return this.start;
    };
}

function waitForElement(elem) {
    if (typeof elem != "undefined" && elem != null) {
        //variable exists, do what you want
    } else {
        setTimeout(waitForElement, 250, elem);
    }
}

function GameObject(sp, args) {

    this.VelocityLimit = 300;
    this.forces = new Array();
    this.phantom = false;
    this.playerId = args.playerId;
    this.playerName = args.playerName;
    this.mass = args.mass;
    this.healthPoints = args.healthPoints;
    this.id = args.id;
    this.ax = 0;
    this.ay = 0;
    this.vx = 0;
    this.vy = 0;
    this.sprite = null;
    this.wm_sprite = null;
    this.spriteType = args.typ;
    this.img = args.img;

    // wait for ID if no ready

    waitForElement(Images);

    if (args.wm_img != null && Images[args.wm_img] != null) {
        this.wm_sprite = new PIXI.Sprite(Images[args.wm_img]);
        this.wm_sprite.anchor.x = 0.5;
        this.wm_sprite.anchor.y = 0.5;
        //        this.wm_sprite.height = args.wm_img.height;
        //        this.wm_sprite.width = args.wm_img.width;
    }

    var txt = Images[args.img];

    if (txt != null && typeof Images != 'undefined') {
        this.sprite = new PIXI.Sprite(Images[args.img]);
    }

    this.prize = "";
    this.prizeValue = "";

    if (this.sprite != null) {
        if ("xPos" in args && "yPos" in args) {
            this.sprite.position.x = args.xPos;
            this.sprite.position.y = args.yPos;
        }
        this.vx = 0;
        this.vy = 0;
        this.sprite.height = args.height;
        this.sprite.width = args.width;
        this.sprite.anchor.x = 0.5;
        this.sprite.anchor.y = 0.5;
    } else {
        console.log("img = " + args.img);
    }

    this.currentAge = 0;
    this.state = args.state;

    this.kind = args.typ & SPRITE_KIND;
    this.typ = args.typ;

    this.shield = null;
    this.shieldOn = false;

    this.phantomCount = 0;
    this.alphaOffset = 0;

    if (this.wm_sprite != null) {
        let height = args.wm_img.height;
        let width = args.wm_img.width;
        theWorldMap.addGameObject(this);
        theWorldMap.setSpritePosition(this.wm_sprite, args.xPos, args.yPos);
    }

    if (this.sprite != null) {
        AppViewport.addChild(this.sprite);
    }

    if (this.kind == gamestate.SpriteKind.PRIZE) {

        var resourceType = args.typ & PRIZE_TYPE;
        var prizeValue = args.typ & PRIZE_VALUE;
        this.prize = resourceType;
        this.prizeValue = prizeValue;

        var label = "";
        if (this.prize == gamestate.PlayerResourceType.SHIELD) {
            label = "S";
        } else if (this.prize == gamestate.PlayerResourceType.LIFE) {
            label = "E";
        } else if (this.prize == gamestate.PlayerResourceType.HYPERDRIVE) {
            label = "H";
        } else if (this.prize == gamestate.PlayerResourceType.BOOSTER) {
            label = "B";
        }

        var style = new PIXI.TextStyle({
            fontFamily: "Arial",
            fontSize: 18,
            fontStyle: "italic",
            stroke: "#334FFF",
            strokeThickness: 2,
            dropShadow: false,
            dropShadowColor: "#000000",
            dropShadowBlur: 2,
            dropShadowAngle: Math.PI / 6,
            dropShadowDistance: 2,
            wordWrap: true,
            wordWrapWidth: 14
        });

        var prizeText = new PIXI.Text(label, style);
        //console.log("Prize Text: " + label + " " + this.prizeValue.toString());
        prizeText.scale.x = 1;
        prizeText.scale.y = 1;

        var str = this.prizeValue.toString(10);
        prizeText.anchor.x = 0.5;
        prizeText.anchor.y = 0.8;

        this.sprite.addChild(prizeText);
    }

    this.GetType = function () {
        return this.spriteType;
    };

    this.phantomCount = 0;
    this.alphaOffset = 0;

    this.setShieldOn = function () {
        if (this.shieldOn == true) {
            return;
        }
        this.shieldOn = true;
        if (this.shield == null) {
            this.shield = new PIXI.Sprite(Images["shield.gif"]);
        }
        this.sprite.addChild(this.shield);
        shieldSound.play();
    };

    this.setShieldOff = function () {
        if (this.shieldOn == false) {
            return;
        }
        this.sprite.removeChild(this.shield);
        this.shieldOn = false;
        shieldSound.stop();
    };

    this.isCloaked = function () {
        var shipState = this.typ & SHIP_STATE;
        return shipState & CLOAK_MODE;
    };

    this.isPhantom = function () {
        var shipState = this.typ & SHIP_STATE;
        return shipState & PHANTOM_MODE;
    };

    this.getKind = function () {
        return this.kind;
    };

    this.isOurShip = function () {
        if (this == OurShip) {
            return true;
        }
        return false;
    };

    this.addForce = function (force) {
        var now = performance.now();
        this.forces.push(force);
    };

    this.rotateTo = function (value) {
        this.sprite.rotation = value;
        if (this.wm_sprite != null) {
            this.wm_sprite.rotation = value;
        }
        return this;
    };

    this.getRotation = function () {
        return this.sprite.rotation;
    };

    this.rotate = function (value) {

        let newR = add(this.sprite.rotation, value);

        let rotation = newR;
        if (newR < 0) {
            rotation = add(newR, TWO_PI);
        } else if (newR > TWO_PI) {
            rotation = subtract(newR, TWO_PI);
        }

        this.rotateTo(rotation);

        return this;
    };

    this.wrap = function () {
        return this;
    };

    this.setPhantom = function (on) {
        if (on) {
            this.phantom = true;
            this.phantomCount = 0;
            this.alphaOffset = 0.0;
        } else {
            this.phantom = false;
        }
    };

    this.setPosition = function (x, y) {
        this.sprite.position.x = x;
        this.sprite.position.y = y;

        if (this.wm_sprite != null) {
            theWorldMap.setSpritePosition(this.wm_sprite, x, y);
        }
    };

    this.move = function (delta, frame) {

        if (this.phantom) {

            if (this.phantomCount % 30 == 0) {
                this.sprite.alpha = .3 + this.alphaOffset;
            } else if (this.phantomCount % 30 == 15) {
                this.sprite.alpha = 0;
                this.alphaOffset += 0.1;
            }
            this.phantomCount += 1;
        } else {
            this.sprite.alpha = 1;
        }

        this.ax = 0;
        this.ay = 0;

        var now = performance.now();
        var survivingForces = new Array();

        for (var i = 0; i < this.forces.length; i++) {

            //          var forceRecord = this.forces[i]
            var force = this.forces[i];
            this.applyForce(force, delta);

            if (!(now - force.getStart() > force.getDuration())) {
                survivingForces.push(force);
            }
        }
        this.forces = survivingForces;

        console.log(sprintf("after applying all forces, accel = %f, %f\n", this.ax, this.ay));

        this.vx = add(this.vx, multiply(this.ax, delta));
        this.vy = add(this.vy, multiply(this.ay, delta));

        this.limitVelocity();
        this.applyDrag(0.1, delta);

        console.log(sprintf("New Velocity = %f, %f\n", this.vx, this.vy));

        let dx = multiply(this.vx, delta);
        let dy = multiply(this.vy, delta);

        console.log(sprintf("Frame %d moving ship by %f, %f\n", frame, dx, dy));

        let x = add(this.sprite.position.x, dx);
        let y = add(this.sprite.position.y, dy);
        this.setPosition(x, y);

        return this;
    };

    this.applyDrag = function (factor, delta) {

        let dvx = multiply(factor, multiply(this.vx, delta));
        this.vx = subtract(this.vx, dvx);

        let dvy = multiply(factor, multiply(this.vy, delta));
        this.vy = subtract(this.vy, dvy);
    };

    this.limitVelocity = function () {
        if (this.vx > this.VelocityLimit) {
            this.vx = this.VelocityLimit;
        } else if (this.vx < -this.VelocityLimit) {
            this.vx = -this.VelocityLimit;
        }
        if (this.vy > this.VelocityLimit) {
            this.vy = this.VelocityLimit;
        } else if (this.vy < -this.VelocityLimit) {
            this.vy = -this.VelocityLimit;
        }
    };

    this.applyForce = function (force, delta) {

        let x = divide(cos(force.getDirection()), this.mass);
        let y = divide(sin(force.getDirection()), this.mass);

        let ax = add(this.ax, multiply(FORCE_COEFFICIENT, multiply(x, force.getMagnitude())));

        let ay = add(this.ay, multiply(FORCE_COEFFICIENT, multiply(y, force.getMagnitude())));

        this.ax = ax;
        this.ay = ay;

    };

    this.distance = function (other) {
        let dX = subtract(this.sprite.position.x, other.sprite.position.x);
        let dY = subtract(this.sprite.position.y, other.sprite.position.y);
        return new Vector(dX, dY);
    };

    this.pullOn = function (other) {
        let dis = other.distance(this);
        let dir = atan2(dis.y, dis.x);


        let F = multiply(FORCE_COEFFICIENT, divide(multiply(GRAVITY, multiply(this.mass, other.mass)), pow(multiply(DISTANCE_COEFFICIENT, dis.Length()), 2)));

        //        console.log("F = " + F);

        let force = new Force(dir, F, 0, 0);
        other.addForce(force);
    };
} // GameObject

function Vector(x, y) {
    this.x = x;
    this.y = y;

    this.Length = function () {
        return sqrt(pow(this.x, 2) + pow(this.y, 2));
    };
}

String.prototype.format = function () {
    var formatted = this;
    for (var arg in arguments) {
        formatted = formatted.replace("{" + arg + "}", arguments[arg]);
    }
    return formatted;
};



function ProgressBar(name, initialValue) {
    this.name = name;

    this.guage1 = document.getElementById(this.name + "-1");
    this.guage2 = document.getElementById(this.name + "-2");
    this.guage3 = document.getElementById(this.name + "-3");
    this.value = initialValue;

    this.update = function (value) {
        this.value = value;
        this.guage3.innerHTML = this.value.toString() + "%";
        this.guage2.setAttribute("style", "width: " + this.value.toString() + "%");
        this.guage1.setAttribute("data-progress", this.value.toString());
    };
}

function RingBuffer(length) {
    /* https://stackoverflow.com/a/4774081 */
    this.insertCursor = 0;
    this.buffer = [];
    this.length = length;

    this.push = function (item) {
        this.buffer[this.insertCursor] = item;
        this.insertCursor = (this.insertCursor + 1) % this.length;
    };

    /**
     * Pulls the last pushed element if back is 0.
     * otherwise pulls the element 'back' slots earlier
     * @param back
     * @returns {*}
     */
    this.pull = function (back) {

        let slot = this.insertCursor - back - 1;
        if (slot < 0) {
            slot = this.length + slot;
        } else if (slot >= this.length) {
            slot = slot - this.length;
        }
        return this.buffer[slot];
    };
};

function keyboard(keyCode) {
    var key = {};
    key.code = keyCode;
    key.isDown = false;
    key.isUp = true;
    key.press = undefined;
    key.release = undefined;
    //The `downHandler`
    key.downHandler = function (event) {
        if (event.keyCode === key.code) {
            if (key.isUp && key.press) key.press();
            key.isDown = true;
            key.isUp = false;
        }
        event.preventDefault();
    };

    //The `upHandler`
    key.upHandler = function (event) {
        if (event.keyCode === key.code) {
            if (key.isDown && key.release) key.release();
            key.isDown = false;
            key.isUp = true;
        }
        event.preventDefault();
    };

    //Attach event listeners
    window.addEventListener("keydown", key.downHandler.bind(key), false);
    window.addEventListener("keyup", key.upHandler.bind(key), false);
    return key;
}

function ClientMessageData(cmd, value) {
    this.cmd = cmd;
    this.value = value;
    this.actionId = ++ActionId;
    LatestUnresolvedActionId = this.actionId;

    this.getCmd = function () {
        return this.cmd;
    };
    this.getValue = function () {
        return this.value;
    };
    this.getActionId = function () {
        return this.actionId;
    };
}
