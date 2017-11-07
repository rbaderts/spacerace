/*
 *
 */
var math = require('mathjs');
var pretty = require('js-object-pretty-print').pretty;
var pb = require('./messages_pb.js');

const COLOR1 = 0x0009FF
const COLOR2 = 0x0FF010
const COLOR3 = 0xFF00B1
const COLOR4 = 0xFFE300
const COLOR5 = 0x00FF0D
const COLOR6 = 0xEEEEEE
const COLOR7 = 0x999999
const COLOR8 = 0xFFFFFF


const SHIP              = 0x010000
const LARGE_ASTEROID    = 0x020000
const SMALL_ASTEROID    = 0x030000
const BULLET            = 0x040000
const BLACKHOLE         = 0x050000
const STAR              = 0x060000
const PRIZE             = 0x070000
const PLANET            = 0x080000

function WorldMap(height, width, wHeight, wWidth) {
    PIXI.Container.call(this);

    this.viewportHeight = height
    this.viewportWidth = width
    this.worldHeight = wHeight
    this.worldWidth = wWidth
    this.GameObjects = {}
}

WorldMap.constructor = WorldMap;
WorldMap.prototype = Object.create(PIXI.Container.prototype);

WorldMap.prototype.addGameObject = function(gameobject) {
    this.GameObjects[gameobject.id] = gameobject
    this.addChild(gameobject.sprite)
}

WorldMap.prototype.removeGameObject = function(gameobject) {
    this.removeChild(gameobject.sprite)
    delete this.GameObjects[gameobject.id]
}


var blinkCount = 0
WorldMap.prototype.computeOverview = function(height, width, scrollX, scrollY) { 
    var graphics = new PIXI.Graphics()

    var xScale = math.divide(width, this.worldWidth)
    var yScale = math.divide(height, this.worldHeight)
    var scrollXScaled = math.multiply(xScale, scrollX)
    var scrollYScaled = math.multiply(yScale, scrollY)
    var xOffset = width / 2
    var yOffset = height / 2

    for (var childIndex in this.GameObjects)  {
        var gameobject = this.GameObjects[childIndex]

        var radius = 2

        if (gameobject.getKind() == SHIP) {
            if (gameobject.isOurShip()) {
                var draw = true

                 if (gameobject.isCloaked()) {
                     if (blinkCount > 16) {
                         blinkCount = 0
                     }
                     if (blinkCount > 8) {
                         draw = false
                     }
                     blinkCount += 1
                 }

                 if (draw) {
         	         graphics.lineStyle(1, COLOR2, 1)
                     radius= 6
                     graphics.beginFill(COLOR6, 1)
                 }
            } else {
                if (!gameobject.isCloaked()) {
         	         graphics.lineStyle(1, COLOR3, 1)
                     radius = 4
                     graphics.beginFill(COLOR6, 1)
                }
            }
        } 
        else if (gameobject.getKind() == STAR) {
             graphics.beginFill(COLOR4, 1)
	         graphics.lineStyle(1, COLOR3, 1)
             radius = 5
        }
        else if (gameobject.getKind() == PLANET) {
             graphics.beginFill(COLOR5, 1)
	         graphics.lineStyle(1, COLOR2, 1)
             radius = 4
        }
        else if (gameobject.getKind() == LARGE_ASTEROID) {
	         graphics.lineStyle(1, COLOR6, 1)
             radius = 3
        }
        else if (gameobject.getKind() == SMALL_ASTEROID) {
	         graphics.lineStyle(1, COLOR6, 1)
             radius = 2
        }
        else if (gameobject.getKind() == PRIZE) {
	         graphics.lineStyle(1, COLOR1, 1)
             radius = 2
        }
        else if (gameobject.getKind() == BLACKHOLE) {
	         graphics.lineStyle(1, COLOR7, 1)
             radius = 4
        }
        else {
             graphics.lineStyle(1, COLOR6, 1)
             radius = 8
        }
        var child = gameobject.sprite
        if (child.position != null) {

            var x = math.multiply(xScale, child.position.x)
            var y = math.multiply(yScale, child.position.y) 

            if (gameobject.getKind() == BULLET) {
	            graphics.lineStyle(1, COLOR3, 1)
                graphics.moveTo(x, y)
                graphics.lineTo(x+1, y+1)
            } else {
                graphics.drawCircle(x, y, radius)
                graphics.endFill()
            }

            if ((gameobject.getKind() == SHIP) && (gameobject.isOurShip())) {
	            graphics.lineStyle(1, COLOR8, 1)
                graphics.moveTo(x, y)
                var x1 = x + 20 * math.cos(gameobject.sprite.rotation)
                var y1 = y + 20 * math.sin(gameobject.sprite.rotation)
	    //       startX := this.Position.x + x*math.Cos(this.Rotation) - y*math.Sin(this.Rotation)
        //       startY := this.Position.y + x*math.Sin(this.Rotation) + y*math.Cos(this.Rotation)
                graphics.lineTo(x1, y1)
            }
        }

    }
    return graphics

}


module.exports = WorldMap;
