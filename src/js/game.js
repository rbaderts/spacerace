var pixi = require('pixi.js');
var pb = require('./messages_pb.js');
var math = require('mathjs');


const Viewport = require('pixi-viewport')
const SettingsPanel = require('settingspanel');
const Debug = require('yy-debug');
const WorldMap = require('./world_map.js');

const SPRITE_KIND  = 0xFF0000
const SHIP_STATE   = 0x0000FF
const PRIZE_TYPE   = 0x00FF00
const PRIZE_VALUE  = 0x0000FF

//const SHIP              = 0x010000
//const LARGE_ASTEROID    = 0x020000
//const SMALL_ASTEROID    = 0x030000
//const BULLET            = 0x040000
//const BLACKHOLE         = 0x050000
//const STAR              = 0x060000
//const PRIZE             = 0x070000
//const PLANET            = 0x080000

const JETS_ON           = 0x000001
const SHIELDS_ACTIVE    = 0x000002
const PHANTOM_MODE      = 0x000004
const CLOAK_MODE        = 0x000008
const TRACTOR_ACTIVE    = 0x000010

const SHIELD            = 0x000100
const BOOSTER           = 0x000200
const HYPERSPACE        = 0x000400
const LIFEENERGY        = 0x000800
const CLOAK             = 0x001000
const TRACTOR           = 0x002000

var pretty = require('js-object-pretty-print').pretty;

//var PIXI = require('pixi');
//var audio = require('pixi-audio'); //pixi-audio is added automatically to the PIXI namespace


var TheGame = null
var GameOver = false
var FreezeDrawing = false
var WS = null
var OurShip = null
var PhysicsUpdateQueue = new Array()
var InventoryUpdateQueue = new Array()
var PlayerUpdateQueue = new Array()
var PlaySoundQueue = new Array()
var OutQueue = new Array()

var theBackground  = null
var theWorldMap  = null

var GameId = document.getElementById("GameID").value
var PlayerId  = document.getElementById("PlayerID").value
var ShipId  = 0
var ActionId  = 0

var TickRate = 30
var READY = false
var SOUNDS_READY = false


var ShieldTokens = 0
var HyperspaceTokens = 0
var BoosterTokens  = 0
var LifeEnergyTokens  = 0
var CloakTokens  = 0
var BlackholeMass = 1000000000000
var BulletSpeed = 60


// Sounds
var thrustSound
var laserSound
var explosionSound
var clickSound
var shieldSound
var boingSound
var bloopSound

const HEIGHT=5000
const WIDTH=5000

const VIEW_HEIGHT=678
const VIEW_WIDTH=678

const OVERVIEW_HEIGHT=288
const OVERVIEW_WIDTH=288

// Scrolling Background

function Background(width, height, x, y) {
  var texture = PIXI.Texture.fromImage('/static/img/background.gif');
  PIXI.extras.TilingSprite.call(this, texture, width, height);

  this.position.x = 0;
  this.position.y = 0;
  this.tilePosition.x = -x  + (VIEW_WIDTH/2)
  this.tilePosition.y = -y + (VIEW_HEIGHT/2)
  this.viewportX = -x + (VIEW_WIDTH/2)
  this.viewportY = -y + (VIEW_HEIGHT/2)
}

Background.prototype = Object.create(PIXI.extras.TilingSprite.prototype);

Background.prototype.scrollTo = function(x, y) {
    this.viewportX = -x - (VIEW_WIDTH/2)
    this.viewportY = -y - (VIEW_HEIGHT/2)
    this.tilePosition.x =  -x  - (VIEW_WIDTH/2)
    this.tilePosition.y =  -y  - (VIEW_HEIGHT/2)
    this.position.x = 0;
    this.position.y = 0;
}


var overviewRenderer = null
var overviewStage = null
var renderer = null
var viewport = null
var stage = null
var graphics = null

function setupWorldMap() {

    overviewRenderer = PIXI.autoDetectRenderer(
      OVERVIEW_HEIGHT, OVERVIEW_WIDTH,
      {antialias: false, transparent: false, resolution: 1}
    );
    overviewRenderer.view.style.border = "border: 2px solid ridge white; padding: 0px;"
    overviewRenderer.backgroundColor = 0x000000;
    document.getElementById("overview_div").appendChild(overviewRenderer.view);
    overviewStage = new PIXI.Container();
    theWorldMap = new WorldMap(VIEW_HEIGHT, VIEW_WIDTH, HEIGHT, WIDTH)
    stage.addChild(theWorldMap)

    graphics = new PIXI.Graphics()
    graphics.x = 0
    graphics.y = 0
    theWorldMap.addChild(graphics)
}


function setupGameArea() {
    //var renderer = new PIXI.CanvasRenderer({width: WIDTH, height: HEIGHT});
    renderer = PIXI.autoDetectRenderer(
      VIEW_HEIGHT, VIEW_WIDTH,
      {antialias: false, transparent: false, resolution: 1}
    );

    renderer.backgroundColor = 0x111111;

    theBackground = new Background(WIDTH , HEIGHT, WIDTH/2, HEIGHT/2)

    stage = new PIXI.Container();

    stage.addChild(theBackground)
    document.getElementById("gamearea_div").appendChild(renderer.view);
    viewport = new Viewport(stage)
    viewport.resize(VIEW_WIDTH, VIEW_HEIGHT, WIDTH, HEIGHT)
}





// DebugPanels


var shipSpritePos = null
var shipPos = null
var viewportCenter = null

function setupDebugPanels() {

    Debug.init();

    shipPos = Debug.add('ShipPos', {text: 'Sprite: 0, 0', side: 'leftBottom'});
    shipSpritePos = Debug.add('ShipSpritePos', {text: 'Sprite: 0, 0', side: 'leftBottom'});
    viewportCenter = Debug.add('ViewportCenter', {text: 'Viewport: 0, 0', side: 'leftBottom'});

    // SettingPanel

    const panel = new SettingsPanel({background: 'rgb(50,50,50)'});

    // create a button that changes its own color on callback and increments a counter on the button
    panel.input('Blackhole Mass',
        function(value)
        {
            var msg = BuildClientMessage(proto.core.CommandType.SETBLACKHOLEMASS, value)
            OutQueue.push(msg)

        }, {original: BlackholeMass, color: 'red'});

    panel.input('Bullet Speed',
        function(value)
        {
            var msg = BuildClientMessage(proto.core.CommandType.SETBULLETSPEED, value)
            OutQueue.push(msg)

        }, {original: BlackholeMass, color: 'red'});

}

var lifeEnergyGauge = document.getElementById("lifeenergy-gauge")

var shieldSpan = document.getElementById("shield-value")
var boostSpan = document.getElementById("boost-value")
var hyperspaceSpan = document.getElementById("hyperspace-value")
var lifeEnergySpan = document.getElementById("lifeenergy-value")
var rotationDisplay = document.getElementById("rotation-display")

function updateGauges() {
    shieldSpan.innerHTML = ShieldTokens.toString()
    boostSpan.innerHTML = BoosterTokens.toString()
    hyperspaceSpan.innerHTML = HyperspaceTokens.toString()
    lifeEnergySpan.innerHTML = LifeEnergyTokens.toString()
    if (OurShip != null) {
        var r = OurShip.sprite.rotation
        rotationDisplay.innerHTML = r.toFixed(3).toString()
    }

    lifeEnergyGauge.setAttribute("aria-valuenow", LifeEnergyTokens.toString())
    lifeEnergyGauge.setAttribute("style", "width: " + LifeEnergyTokens.toString()+ "%;")


}

/*
PIXI.loader.add([ 
    "/static/img/smallroid_2.gif",
    "/static/img/largeroid_2.gif",
    "/static/img/bullet.gif",
    "/static/img/blackhole.gif",
    "/static/img/star.gif",
    "/static/img/planet.gif",
    "/static/img/background.gif",
    "/static/img/SS.gif",
    "/static/img/SWS.gif",
    "/static/img/SSJ.gif",
    "/static/img/tractor.gif",
    "/static/img/prize.gif"]).load(setup);
    */



PIXI.loader.add("/static/img/sheet.json").load(setup);
var ID = null


sounds.load([
    "/static/snd/thrust2.wav",
    "/static/snd/laser.wav",
    "/static/snd/explosion1.wav",
    "/static/snd/click.wav",
    "/static/snd/shield.wav",
    "/static/snd/boing.wav",
    "/static/snd/bloop.wav"
]);

function round(num) {
     return (Math.round(num*1000))/1000
}


function setupSounds() {
    thrustSound = sounds["/static/snd/thrust2.wav"]
    laserSound = sounds["/static/snd/laser.wav"]
    explosionSound = sounds["/static/snd/explosion1.wav"]
    clickSound = sounds["/static/snd/click.wav"]
    shieldSound = sounds["/static/snd/shield.wav"]
    boingSound = sounds["/static/snd/boing.wav"]
    bloopSound = sounds["/static/snd/bloop.wav"]
    SOUNDS_READY = true
}

sounds.whenLoaded = setupSounds;

function dynamicCall (func) {
    this[func].apply(this, Array.prototype.slice.call(arguments, 1));
}

function setup() {
    ID = PIXI.loader.resources["/static/img/sheet.json"].textures;
    setupGameArea()
    setupWorldMap()
    setupDebugPanels()
    renderer.render(stage)
    overviewRenderer.render(overviewStage)
    startGame()
    READY = true
    requestAnimationFrame(gameLoop);
    setupDebugPanels()
}



var frameOffset = -1
var counter1 = 0

function bulletSpeedChange(element) {
   var speed = element.val()
   var msg = BuildClientMessage(proto.core.CommandType.SETBULLETSPEED, speed)
   OutQueue.push(msg)
}


function Now() {
    return performance.now() 
}


var T = 0
var LastFrame = 0
var Frame = 0


function gameLoop (timestamp) {

    var fps = 40

        if (Frame % 60 == 0) {
            updateGauges()
        }
        var n = Now()
        if (T === 0) {
            T = timestamp
            LastFrame = timestamp
        }

        var frameTime = timestamp - LastFrame  // get the delta time since last frame
        LastFrame = timestamp

        if (Frame%100 === 0) {
            console.log("Frame: " + Frame + ", timestamp = " + timestamp + ", T = " + T)
        }

        if (FreezeDrawing == false) {
            update(frameTime / 1000)
            updateOverview()

            draw()
        }
        T += frameTime
        Frame += 1


        sendUpdates(frameTime)

        updateDebugs()
        
        requestAnimationFrame(gameLoop)
}


function updateOverview() {
    
    ov = theWorldMap.computeOverview(OVERVIEW_HEIGHT, OVERVIEW_WIDTH, 
                                    viewport.center.x, viewport.center.y) 
    overviewStage.removeChildren()
    overviewStage.addChild(ov)

}


var DebugCount = 0;
function updateDebugs() {

    if (DebugCount != 20) {
        DebugCount += 1
        return
    }
    DebugCount = 0

    if (OurShip != null) {
        var pos = "Ship: (" + OurShip.x.toFixed(2).toString() + ", " + OurShip.y.toFixed(2).toString() + ")";
        Debug.one(pos, {panel: shipPos, color: 'red'});
    }
    if (OurShip != null) {
        var pos = "Sprite: (" + OurShip.sprite.x.toFixed(2).toString() + ", " + OurShip.sprite.y.toFixed(2).toString() + ")";
        Debug.one(pos, {panel: shipSpritePos, color: 'red'});
    }
    if (viewport != null) {
        var pos = "Viewport Center: (" + viewport.center.x.toFixed(2) + ", " + 
            viewport.center.y.toFixed(2) + ")"
        Debug.one(pos, {panel: viewportCenter, color: 'red'});
    }
}



function sendUpdates(delta) {

    if (OutQueue.length > 0) {
         var msgs = new(proto.core.ClientMessages)
         msgs.setMessagesList (OutQueue)
         var b = msgs.serializeBinary()

//          proto.core.ServerMessage.InitializePlayerData.prototype.serializeBinary = function() {
 //        var b = proto.core.ServerMessage.serializeBinary(msg.data)
         send(b)
         //WS.send(b);
         OutQueue = new Array()
    }

}

function send(data) {
    if (WS != null) {
         WS.send(data)
    }
}

function update(delta) {

    if (counter1 == 5) {
        counter1 = 0
        if (leftArrow.isDown) {
           OurShip.rotate(-0.1)
           var msg = BuildClientMessage(proto.core.CommandType.ROTATE, -0.1)
           OutQueue.push(msg)
        } 
        else if (rightArrow.isDown) {
           OurShip.rotate(0.1)
           var msg = BuildClientMessage(proto.core.CommandType.ROTATE, 0.1)
           OutQueue.push(msg)
        } 

    }

    counter1 += 1

    var msgs = PhysicsUpdateQueue.length
    var highestFrame = 0
    var updateMessage = null
    
    for (i = 0; i < msgs; i++) {
        if (PhysicsUpdateQueue[i].getFrame() > highestFrame) {
            updateMessage = PhysicsUpdateQueue[i]
            highestFrame = updateMessage.getFrame()
        }
    }

    if (msgs >  1) {
         console.log("dropping " + (msgs-1) + " extra server updates")
    }

    if (frameOffset === -1) {
        frameOffset = highestFrame - Frame
        console.log("FrameOffset = " + frameOffset)
    }

    if (Frame % 60 === 0) {
        console.log("Frame: " + Frame + ", serverFrame: " + highestFrame) 
    }
    
    PhysicsUpdateQueue = new Array()


    var playerUpdates = PlayerUpdateQueue.length
    for (i = 0; i < playerUpdates; i++) {
        processPlayerUpdate(PlayerUpdateQueue[i])
    }
    updateGauges()
    PlayerUpdateQueue = new Array()
    
    // Calculate local physics

    if (OurShip != null) {
        OurShip.move(delta)
        OurShip.sprite.x = OurShip.x
        OurShip.sprite.y = OurShip.y
 
        if (viewport != null) {
            viewport.moveCenter(OurShip.x, OurShip.y)
//            theBackground.scrollTo(-OurShip.x, -OurShip.y)
        }
    }

    // Play Sounds
    
    var sounds = PlaySoundQueue.length
    for (i = 0; i < sounds; i++) {
        handlePlaySound(PlaySoundQueue[i])
    }
    PlaySoundQueue = new Array()

//    for (var s in TheGame.all) {
//        var sp = TheGame.all[s]
//        sp.move(delta).wrap()
//    }

 //   scrollBackground()

    viewport.update()
    // Adjust 
    if (updateMessage != null) {
        updateSprites(updateMessage)
    } 
}

function processPlayerUpdate(update) {

    console.log("ProcessPlayerUpdate:  " + JSON.stringify(update))
    var player = update.getPlayer()

    var inventories = player.getInventoryList()
    var len = inventories.length;

    for (var i = 0; i < len; i++) {
        var inventory = inventories[i]

        if (inventory.getResourcetype() == proto.core.PlayerResourceType.SHIELDRESOURCE) {
            ShieldTokens = inventory.getValue()
        } 
        else if (inventory.getResourcetype() == proto.core.PlayerResourceType.BOOSTERRESOURCE) {
            BoosterTokens = inventory.getValue()
        } 
        else if (inventory.getResourcetype() == proto.core.PlayerResourceType.HYPERSPACERESOURCE) {
            HyperspaceTokens = inventory.getValue()
        } 
        else if (inventory.getResourcetype() == proto.core.PlayerResourceType.LIFEENERGYRESOURCE) {
            LifeEnergyTokens = inventory.getValue()
        } 
        else if (inventory.getResourcetype() == proto.core.PlayerResourceType.CLOAKRESOURCE) {
            CloakTokens = inventory.getValue()
        } 
    }

}

function scrollBackground() {
}

function draw() {

    renderer.render(stage)
    overviewRenderer.render(overviewStage)
}





function updateSprites(updateMessage, delta) {


    var actionId = updateMessage.getActionid();
    var spritelist = {}
    for (var s in TheGame.all) {
         spritelist[s] = true
    }

    var sprites = updateMessage.getSpritesList()
    var len = sprites.length;

    for (var i = 0; i < len; i++) {
        var sprite = sprites[i]
//        console.log(JSON.stringify(sprite))
        TheGame.updateSprite(sprite, actionId)
        delete spritelist[sprite.getId()]
    }
    for (var s in spritelist) {
        console.log("Deleting sprite id: " + s)
        var sprite = TheGame.all[s]
        theWorldMap.removeGameObject(sprite)
        delete TheGame.all[s]
    }

}

jQuery('#newgame').on('click', function() {
    jQuery.get( "/newgame", function( data ) {
        startGame()
    });
});


function startGame() {
    runGame()
}


function runGame() {

        TheGame = new Game()
        TheGame.init()

        WS = new WebSocket('ws://' + window.location.host + '/updates/' + GameId )
        WS.binaryType = 'arraybuffer';

        WS.onopen = function () {
            console.log("Server connection enstablished");
        }

        WS.onclose = function () {
            console.log("Server close");
            WS = null
            var page = "/lobby"
            window.location.assign(page)
            //Server = null
        }

        WS.onerror = function (error) {
            console.log("Server error: " + error);
            WS = null
            var page = "/lobby"
            window.location.assign(page)
        }

        WS.onmessage = function(msg) {

            if (TheGame == null) {
                return
            }
            handleServerCommand(msg)
        }
}


function  handleServerCommand(msg) {

    var message = proto.core.ServerMessage.deserializeBinary(msg.data)
    if (message.getTyp() === proto.core.MessageType.PHYSICSUPDATE) {
        PhysicsUpdateQueue.push(message.getUpdate())
    } 
    else if (message.getTyp() === proto.core.MessageType.PLAYERUPDATE) {
        PlayerUpdateQueue.push(message.getPlayers())
    //else if (message.getTyp() === proto.core.MessageType.INVENTORYUPDATE) {
    //    InventoryUpdateQueue.push(message.getInventory())
    //} 
    } 
   else if (message.getTyp() === proto.core.MessageType.PLAYERINITIALIZE) {
        handlePlayerInitialize(message.getInitialize())
    } 
    else if (message.getTyp() === proto.core.MessageType.PLAYSOUND) {
        PlaySoundQueue.push(message.getSound())
    } 
    else if (message.getTyp() === proto.core.MessageType.DRAWMESSAGE) {
        handleDrawMessage(message.getDraw())
    } 
    else if (message.getTyp() === proto.core.MessageType.FREEZEDRAWING) {
        FreezeDrawing = true
    } 
    else if (message.getTyp() === proto.core.MessageType.PLAYERDEAD) {
        handlePlayerDead(message.getDead())
    } 
    else {
        console.log("Received server message type: " + message.typ)
        console.log("message = " + JSON.stringify(message))
        console.log("msg.data = " + JSON.stringify(msg.data))
    }


}

function handlePlayerInitialize(p) {
    PlayerId = p.getPlayerid()
    console.log("init: PlayerId = " +PlayerId)
}

function handlePlaySound(sound) {

    //console.log("handlePlaySound")
    //console.log(JSON.stringify(sound))
    if (sound.getSoundtype() == proto.core.SoundType.EXPLOSIONSOUND) {
       explosionSound.volume = sound.getVolume()
       explosionSound.restart()
    } else if (sound.getSoundtype() == proto.core.SoundType.BOINGSOUND) {
       boingSound.volume = sound.getVolume()
       boingSound.restart()
    } else if (sound.getSoundtype() == proto.core.SoundType.BLOOPSOUND) {
       bloopSound.volume = sound.getVolume()
       bloopSound.restart()
    }

}

function handlePlayerDead(dead) {
    if (dead.getPlayerid() == PlayerId) {
        var labelStyle = new PIXI.TextStyle({fontFamily : '"Trebuchet MS', fontStyle: "italic", fontSize: 42, stroke : 0xedffff, stokeThickness : 2, align : 'center'});


        let deadText = new PIXI.Text("You Are Dead - You Suck !!!",  labelStyle);
        deadText.anchor.x = 0.5
        deadText.anchor.y = 0.5
        theWorldMap.addChild(deadText);
        renderer.render(stage)
    }
}


function eraseTractor(ship) {
    graphics.clear()
}

function drawTractor(ship) {

    console.log("draw tractor")

	var x = math.divide(ship.sprite.width, 2)
	var y = 0

	var sourceX = ship.sprite.x + (math.multiply(x, math.cos(ship.sprite.rotation))) - (math.multiply(y, math.sin(ship.sprite.rotation)))
	var sourceY = ship.sprite.y + (math.multiply(x, math.sin(ship.sprite.rotation))) + (math.multiply(y,Math.cos(ship.sprite.rotation)))

    var destX = sourceX + math.multiply(20, math.cos(ship.sprite.rotation))
    var destY = sourceY + math.multiply(20, math.sin(ship.sprite.rotation))

    graphics.lineStyle(2, 0xccff10, 1)
    graphics.moveTo(sourceX, sourceY)
    graphics.lineTo(destX, destY)

}

function handleDrawMessage(draw) {

    var graphics = new PIXI.Graphics()
    var cmds = draw.getCmdsList();
    var len = cmds.length;

    for (var i = 0; i < len; i++) {
        console.log("cmd = " +cmds[i])
        eval("graphics." + cmds[i])
    }
    graphics.x = 0
    graphics.y = 0
    theWorldMap.addChild(graphics)
    renderer.render(stage)

}


function hasState(stateList, state) {
    var len = stateList.length;
    for (var i = 0; i < len; i++) {
        if (stateList[i] == state) {
            return true
        }
     }
     return false
}

function hasProperty(propertyList, prop) {
    var len = propertyList.length;
    for (var i = 0; i < len; i++) {
        if (propertyList[i] == prop) {
            return true
        }
     }
     return false
}


var leftArrow = keyboard(37)
var upArrow = keyboard(38)
var rightArrow = keyboard(39)
var downArrow = keyboard(40)
var space = keyboard(32)
var esc = keyboard(27)
var tab = keyboard(9)
var control = keyboard(17)
var s = keyboard(83)
var w = keyboard(87)
var c = keyboard(67)
var t = keyboard(84)

function BuildClientMessage(c, v) {

   var actionId = ++ActionId
   var cmd = new proto.core.PlayerCommandMessage()

   cmd.setCmd(c)
   cmd.setValue(v)
   cmd.setActionid(actionId)

   var msg = new proto.core.ClientMessage()

   msg.setTyp(proto.core.MessageType.PLAYERCOMMAND)
   msg.setCmd(cmd)

   return msg

}
function Game() {


    this.state = this.play
    //var all = new Set();
    this.all = {}

    this.init = function() {


       esc.press = function() {
       };

       leftArrow.press = function() {

           var msg = BuildClientMessage(proto.core.CommandType.ROTATE, -0.1)

           OurShip.rotate(-0.1)
           OutQueue.push(msg)
       };

       rightArrow.press = function() {
           var msg = BuildClientMessage(proto.core.CommandType.ROTATE, 0.1)
           OurShip.rotate(0.1)
           OutQueue.push(msg)
       };

       upArrow.press = function() {

           var msg = BuildClientMessage(proto.core.CommandType.THRUST, 40)

           var actionId = msg.getCmd().getActionid()
           var force = new Force(OurShip.sprite.rotation, 2000, actionId)
           OurShip.addForce(force, 100)

           OutQueue.push(msg)
           thrustSound.restart()
       };

       tab.press = function() {
           if (BoosterTokens >= 1) {
                 var msg = BuildClientMessage(proto.core.CommandType.BOOSTER, 120)
                 var actionId = msg.getCmd().getActionid()
                 var force = new Force(OurShip.sprite.rotation, 2000, actionId)
                 OurShip.addForce(force, 100)

                 OutQueue.push(msg)
                 thrustSound.volume=0.7
                 thrustSound.restart()
                 thrustSound.volume=0.5
             }

       };

       s.press = function() {
           if (ShieldTokens > 1) {
              var msg = BuildClientMessage(proto.core.CommandType.SHIELDON, 5)
              OutQueue.push(msg)
              shieldSound.restart()
           }
       };

       s.release = function() {
           var msg = BuildClientMessage(proto.core.CommandType.SHIELDOFF, 5)
           OutQueue.push(msg)
           shieldSound.pause()
       };

       w.press = function() {
           var msg = BuildClientMessage(proto.core.CommandType.HYPERSPACE, 5)
           OutQueue.push(msg)
       };

       c.press = function() {
           var msg = BuildClientMessage(proto.core.CommandType.CLOAK, 5)
           OutQueue.push(msg)
       };

       t.press = function() {
           var msg = BuildClientMessage(proto.core.CommandType.TRACTORON, 5)
           OutQueue.push(msg)
       };

       t.release = function() {
           var msg = BuildClientMessage(proto.core.CommandType.TRACTOROFF, 5)
           OutQueue.push(msg)
       };


       downArrow.press = function() {
       };


       space.press = function() {
           //this.fire()
//           if (control.isDown) {
//               var cmd = {cmd: "Phaser", val: 60}
 //              Server.send(JSON.stringify(cmd));
 //              laser.restart()
 //          } else {
               var msg = BuildClientMessage(proto.core.CommandType.FIRE, 400)
               OutQueue.push(msg)
               laserSound.restart()
  //         }
       };


       //this.T_ship = new GameObject({name: "Ship", maxVelocity: 150,  img: "/static/img/ship.gif", xPos: 200, yPos: 200, vx: 0, vy: 0})
       //this.all.add(T_ship)

       //this.newRoid()
       //this.newRoid()
       //this.newBlackhole()

       //this.gameLoop()
    }


    this.updateSprite = function(sprite, actionId) {


        if (!(sprite.getId() in this.all)) {
           //console.log("addingSprite: " + sprite.getId())

           var img = null
           console.log(pretty(sprite))
           console.log("typ = " + sprite.getTyp().toString(16))
           //var typInfo  = parseInt(sprite.getTyp(), 16)
           console.log(sprite.getTyp().toString(16) + " && " + SPRITE_KIND.toString(16))
           var typ = sprite.getTyp() & SPRITE_KIND

           console.log("kind = " + typ.toString(16))
           console.log("kind(10) = " + typ.toString(10))

           if (typ == proto.core.SpriteType.LARGEASTEROID) {
                //img = "/static/img/largeroid_2.gif"
                img = "largeroid_2.gif"
           } 
           else if (typ == proto.core.SpriteType.SMALLASTEROID) {
                //img = "/static/img/smallroid_2.gif"
                img = "smallroid_2.gif"
           } 
           else if (typ == proto.core.SpriteType.SHIP) {
                //img = "/static/img/SS.gif"
                img = "SS.gif"
           } 
           else if (typ == proto.core.SpriteType.BULLET) {
                //img = "/static/img/bullet.gif"
                img = "bullet.gif"
           } 
           else if (typ == proto.core.SpriteType.BLACKHOLE) {
                //img = "/static/img/blackhole.gif"
                img = "blackhole.gif"
           } 
           else if (typ == proto.core.SpriteType.STAR) {
                //img = "/static/img/star.gif"
                img = "star.gif"
           }
           else if (typ == proto.core.SpriteType.PRIZE) {
                //img = "/static/img/prize.gif"
                img = "prize.gif"
           }
           else if (typ == proto.core.SpriteType.PLANET) {
                //img = "/static/img/planet.gif"
                img = "planet.gif"
           }
           else {
               console.log("Unknown sprite type: " + typ) 
           }
            
            //s = new PIXI.Sprite(PIXI.loader.resources[img].texture)

            var s = new GameObject(sprite, {typ: sprite.getTyp(), id: sprite.getId(), img: img,
                                   height: sprite.getHeight(), width: sprite.getWidth(), xPos: sprite.getX(), 
                                   yPos: sprite.getY(), vx: sprite.getVx(), vy: sprite.getVy(), mass: sprite.getMass(), maxAge: 1000, 
                                   state: proto.core.SpriteStatus.NOSTATE})

            if ((typ === proto.core.SpriteType.SHIP) && (sprite.getPlayerid() === PlayerId)) {

                s.sprite.position.x = s.x
                s.sprite.position.y = s.y

                OurShip = s
                if (viewport != null) {
                    //theScroller.scrollTo(OurShip.x, OurShip.y)
                    viewport.moveCenter(OurShip.x, OurShip.y)
//                    theBackground.scrollTo(-OurShip.x, -OurShip.y)
                }
            }

            this.all[sprite.getId()] = s

        } else {

            var s = this.all[sprite.getId()]
            var ignoreShipUpdates = false
            var ps = s.sprite


           s.typ = sprite.getTyp()
           if (s != OurShip) {

                ps.position.set(sprite.getX(), sprite.getY())
               //
//                ps.position.set(sprite.getX() - theScroller.getXOffset(), sprite.getY() - theScroller.getYOffset())
                //ps.position.set(sprite.getX(), sprite.getY() )

                console.log("updateSprite: type="+s.typ+", old rotation = "+ s.sprite.rotation + ", new rotation = " + sprite.getRotation()) 
                s.rotateTo(sprite.getRotation())

           } else if (s === OurShip) {

//////////////////                var deltaX = round(ps.position.x - sprite.getX())
//                var deltaY = round(ps.position.y - sprite.getY())
                var deltaX = round(s.x - sprite.getX())
                var deltaY = round(s.y - sprite.getY())
                console.log("ship sprite: " + sprite.getId() + ", delta: (" + deltaX + "," + deltaY + ")")

                var dis = Math.sqrt(Math.pow(deltaX, 2) + Math.pow(deltaY, 2));
                dis = round(dis)
                if (dis > 2) {
//                    ps.position.set(sprite.getX(), sprite.getY())
//                    ps.position.set(sprite.getX(), sprite.getY())
                    s.x = sprite.getX()
                    s.y = sprite.getY()
                    console.log("yanking sprite " + sprite.getId() + " into position")
                } else if (dis > 0.1) {
                    s.x = round(s.x - (deltaX * 0.1))
                    s.y = round(s.y - (deltaY * 0.1))
                }

                console.log("ship.x = " + s.x + ", ship.y: " + s.y)
                console.log("ship.sprite.x = " + s.sprite.x + ", ship.sprite.y: " + s.sprite.y)

                console.log("ourShip: old rotation = "+ ps.rotation + ", new rotation = " + sprite.getRotation()) 

                var deltaRotation = Math.abs(ps.rotation - sprite.getRotation())
                if (deltaRotation > 0.2) {
                    ps.rotation = sprite.getRotation()
                } else if (deltaRotation > 0.1) {
                    if (sprite.getRotation() > ps.rotation) {
                        ps.rotation += 0.1
                    } else {
                        ps.rotation += -0.1
                    }
                }

                ps.vx = sprite.getVx()
                ps.vy = sprite.getVy()


           }

           if (s.getKind() == proto.core.SpriteType.SHIP) {

                var img = null

                var shipState = sprite.getTyp() & SHIP_STATE
                var shield  = shipState & SHIELDS_ACTIVE
                var jets = shipState & JETS_ON
                var phantom = shipState & PHANTOM_MODE
                var tractor = shipState & TRACTOR_ACTIVE

                if (s.phantom != phantom) {
                    s.setPhantom(phantom)
                }

                if (shield) {
                    //img = "/static/img/SWS.gif"
                    img = "SWS.gif"
                } else if (jets) {
                    //img = "/static/img/SSJ.gif"
                    img = "SSJ.gif"
                } else {
                    //img = "/static/img/SS.gif"
                    img = "SS.gif"
                }

                if (img != s.img) {
                    s.img = img
                    //ps.texture = PIXI.loader.resources[img].texture
                    if (ID != null) {
                        ps.texture = ID[img]
                    }

                    ps.height = ps.texture.height
                    ps.width = ps.texture.width
                }

                if (tractor) {
                    drawTractor(s)
                }

            }
        }

    }

    this.play = function() {
    }

    this.keyboard = function(keyCode) {
      var key = {};
      key.code = keyCode;
      key.isDown = false;
      key.isUp = true;
      key.press = undefined;
      key.release = undefined;
      //The `downHandler`
      key.downHandler = function(event) {
        if (event.keyCode === key.code) {
          if (key.isUp && key.press) { 
              key.press();
          } else if (event.repeat == true) {
              key.press()
          }
          key.isDown = true;
          key.isUp = false;
        }
        event.preventDefault();
      };
      //The `upHandler`
      key.upHandler = function(event) {
        if (event.keyCode === key.code) {
          if (key.isDown && key.release) key.release();
          key.isDown = false;
          key.isUp = true;
        }
        event.preventDefault();
      };
      //Attach event listeners
      window.addEventListener(
        "keydown", key.downHandler.bind(key), false
      );
      window.addEventListener(
        "keyup", key.upHandler.bind(key), false
      );
      return key;
    }

}

var FRAME_RATE_NANOS=16670000


function Force(d, m, actionId) {

    this.direction = d
    this.magnitude = m
    this.actionId = actionId

    this.getMagnitude = function() {
        return this.magnitude
    }
    this.getDirection = function() {
        return this.direction
    }

    this.getActionId = function() {
        return this.actionId
    }

}

function GameObject(sp, args) {


    this.forces = {}
    this.forceCount =0
    this.phantom = false
    this.cloaked = false

    this.mass = 1
    if ('mass' in args) {
        this.mass = args.mass
    }

    this.id = args.id

    this.ax = 0
    this.ay = 0

    //this.sprite = new PIXI.Sprite(PIXI.loader.resources[args.img].texture)
    if (ID != null) {
        this.sprite = new PIXI.Sprite(ID[args.img]);
    }

    this.prize = ""
    this.prizeValue = ""

    this.x = 0
    this.y = 0
    this.sprite.x = 0
    this.sprite.y = 0
    this.sprite.vx = 0
    this.sprite.vy = 0
    this.sprite.height = args.height
    this.sprite.width = args.width
    this.sprite.anchor.x = 0.5
    this.sprite.anchor.y = 0.5
//    this.sprite.rotation = math.divide (math.PI, 2)
    this.currentAge = 0
    this.state = args.state

    this.kind = args.typ & SPRITE_KIND
    this.typ = args.typ

    theWorldMap.addGameObject(this)

    if (this.kind == proto.core.SpriteType.PRIZE) {

        var prizeType = args.typ & PRIZE_TYPE
        var prizeValue = args.typ & PRIZE_VALUE
        this.prize = prizeType
        this.prizeValue = prizeValue

        var label = ""
        if (this.prize == SHIELD) {
            label = "S"
        }
        else if (this.prize == LIFEENERGY) {
            label = "E"
        }
        else if (this.prize == HYPERSPACE) {
            label = "H"
        }
        else if (this.prize == BOOSTER) {
            label = "B"
        }

        var labelStyle = new PIXI.TextStyle({fontFamily : '"Trebuchet MS', fontStyle: "italic", fontSize: 22, stroke : 0xedffff, stokeThickness : 2, align : 'center'});


        let prizeText = new PIXI.Text(label + "\n" + this.prizeValue,  labelStyle);
        prizeText.anchor.x = 0.5
        prizeText.anchor.y = 0.5

        this.sprite.addChild(prizeText)

    }

    this.phantomCount = 0
    this.alphaOffset = 0

    if ('vx' in args) {
        this.sprite.vx = args.vx
    }
    if ('vy' in args) {
        this.sprite.vy = args.vy
    }

    if ('xPos' in args) {
        this.x = args.xPos
        this.sprite.x = args.xPos
    }
    if ('yPos' in args) {
        this.y = args.yPos
        this.sprite.y = args.yPos
    }
  
    this.isCloaked = function() {
        var shipState = this.typ & SHIP_STATE
        var cloaked = shipState & CLOAK_MODE
        return cloaked
    }

    this.getKind = function() {
        return this.kind
    }

    this.isOurShip = function() {
        if (this == OurShip) {
            return true
        }
        return false
    }

    this.addForce = function(force, duration) {
       var now = performance.now() 
       this.forceCount += 1
       var str = this.forceCount.toString(10)
       this.forces[str] = {start: now, duration: duration, force: force}
    }
   
    this.rotateTo = function(value) { 
       this.sprite.rotation = value
       return this
    }

    this.rotate = function(value) { 
        var twoPi = math.multiply(math.PI, 2)
        var newR = math.add(this.sprite.rotation, value)
        if (newR < 0) {
            this.sprite.rotation = math.add(twoPi, newR)
        } else if (newR > twoPi) {
            this.sprite.rotation = math.subtract(newR, twoPi)
        } else {
            this.sprite.rotation = newR
        }
        return this
    }

    this.wrap = function() { 

        /*
       if (this.sprite.x <= 0) { 
           this.sprite.x = WIDTH
       } 
       else if (this.sprite.x >= WIDTH)  {
           this.sprite.x = 0
       }

       if (this.sprite.y <= 0) { 
           this.sprite.y = HEIGHT
       } 
       else if (this.sprite.y >= HEIGHT)  {
           this.sprite.y = 0
       }

      */
 
       return this
    }
 
    this.setPhantom = function(on) {
        if (on) {
            this.phantom = true
            this.phantomCount = 0
            this.alphaOffset = 0.0
        }
        else {
            this.phantom = false
        }
    }

    var moveCount = 0
    this.move = function(delta) {

        if (this.phantom) {

           if (this.phantomCount % 30 == 0) {
               this.sprite.alpha = .3 + this.alphaOffset
           } else if (this.phantomCount % 30 == 15) {
               this.sprite.alpha = 0
               this.alphaOffset += 0.1
           }
           this.phantomCount += 1

        } else {
           this.sprite.alpha = 1
        }

        this.ax = 0
        this.ay = 0

        var now = performance.now() 

        for (var f in this.forces)  {
            var forceRecord = this.forces[f]
            var force = forceRecord.force
            this.applyForce(force, delta)
            if (now - forceRecord.start > forceRecord.duration) {
                delete this.forces[f]
            } 
        }

        var dx = 0
        var dy = 0

        if (!math.isZero(this.ax)) {
            this.sprite.vx += math.multiply(this.ax, delta)
        }
        dx = math.multiply(this.sprite.vx, delta)

        if (!math.isZero(this.ay)) {
            this.sprite.vy += math.multiply(this.ay, delta)
        }
        dy = math.multiply(this.sprite.vy, delta)

        //this.applyDrag(.10, delta)

        //this.sprite.x = math.add(this.sprite.x, dx)
        //this.sprite.y = math.add(this.sprite.y, dy)
        this.x = math.add(this.x, dx)
        this.y = math.add(this.y, dy)

        if (moveCount == 20) {
            console.log("ship x,y : " + this.x + "," + this.y)
            console.log("ship.sprite x,y : " + this.sprite.x + "," + this.sprite.y)
            moveCount = 0
        }
        moveCount += 1
        return this
    }

    this.applyForce = function(force, delta) {

        // F = m * a :     a = F/m
        var quarterCircle = math.divide(math.PI, 2)

        var ignoreX = false
        var ignoreY = false
        if (math.isZero(math.mod(force.direction, math.multiply(quarterCircle, 2)))) {
            ignoreY = true
        } 
        else if (math.isZero(math.mod(force.direction,quarterCircle))) {
            ignoreX = true
        }
        
        if (!ignoreX) {
            var x = math.cos(force.direction) / this.mass
            this.ax += round(x * force.magnitude)
        }
        if (!ignoreY) { 
            var y = math.sin(force.direction) / this.mass
            this.ay += round(y * force.magnitude)
        }

        if (!ignoreX && !ignoreY) {
            console.log("applyForce:  ax: " + this.ax + ", ay: " + this.ay)
        }
        else if (!ignoreX) {
            console.log("applyForce:  ax: " + this.ax + ", ZERO")
        }
        else {
            console.log("applyForce:  ZERO, ay: " + this.ay)
        }


 
    }


}
  


String.prototype.format = function() {
    var formatted = this;
    for( var arg in arguments ) {
        formatted = formatted.replace("{" + arg + "}", arguments[arg]);
    }
    return formatted;
};



/*

function Scroller(stage) {
  this.X = 0
  this.Y = 0
  this.stage = stage

  this.background = new Background(1366, 768, 0, 0);
  stage.addChild(this.background);

  theWorldMap = new GameMap(VIEW_HEIGHT, VIEW_WIDTH, HEIGHT, WIDTH)
  stage.addChild(theWorldMap)
//  this.X = HEIGHT/2
//  this.Y = WIDTH/2

}

Scroller.prototype.scrollTo = function(x, y) { 
    this.X = x
    this.Y = y
    this.background.scrollTo(x, y)
    theWorldMap.scrollTo(x, y)
};

Scroller.prototype.getXOffset = function() {
    return -this.background.viewportX
}
Scroller.prototype.getYOffset = function() {
    return -this.background.viewportY
}

/*

Scroller.prototype.setViewport = function(x, y) {
  //this.background.update();
  this.background.setViewport(x, y)
  this.tilePosition.x = 0;
  this.tilePosition.y = 0;
};
*/
/*
Scroller.prototype.scroll = function(x, y) 
{ 
    this.background.scroll(x, y);
//  this.mid.setViewportX(viewportX);
};
*/

