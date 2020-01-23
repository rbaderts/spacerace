/*
 *
 */
import * as PIXI from 'pixi.js';

import {
    divide, multiply, evaluate
} from 'mathjs'


var pretty = require('js-object-pretty-print').pretty;
//var pb = require('./messages_pb.js');
var gamestate = require('./gamestate_generated.js').messages;

const COLOR1 = 0x0009FF;
const COLOR2 = 0x0FF010;
const COLOR3 = 0xF91710;
const COLOR4 = 0xFFE300;
const COLOR5 = 0x00FF0D;
const COLOR6 = 0xF9F210;
const COLOR7 = 0x544A92;
const COLOR8 = 0xFFFFFF;
const COLOR9 = 0xD5310D;
const COLOR10 = 0x163EC8;

const SHIPICON = "shipicon.gif";
function WorldMap(viewport, height, width, wHeight, wWidth) {
    PIXI.Container.call(this);

    this.viewportHeight = height;
    this.viewportWidth = width;
    this.worldHeight = wHeight;
    this.worldWidth = wWidth;
    this.GameObjects = new Map();
    this.XScale = divide(width, this.worldWidth);
    this.YScale = divide(height, this.worldHeight);
    this.Viewport = viewport;
}

WorldMap.constructor = WorldMap;
WorldMap.prototype = Object.create(PIXI.Container.prototype);

WorldMap.prototype.addChildSprite = function (sprite) {
    //    sprite.height = this.XScale * height;
    //    sprite.width = this.XScale * width;
    this.Viewport.addChild(sprite);
};

WorldMap.prototype.addGameObject = function (gameobject) {
    //    console.log("adding to worldmap id : " + gameobject.id);
    //    this.GameObjects.set(gameobject.id, gameobject);
    this.addChildSprite(gameobject.wm_sprite);

    console.log("WorldMap addGameObject: with size h = " + gameobject.wm_sprite.height + ", w = " + gameobject.wm_sprite.width);
};

WorldMap.prototype.removeGameObject = function (gameobject) {
    this.removeChild(gameobject.wm_sprite);
    //    this.GameObjects.delete(gameobject.id)
};

WorldMap.prototype.setSpritePosition = function (sprite, x, y) {

    let x1 = multiply(this.XScale, x);
    let y1 = multiply(this.YScale, y);

    sprite.position.x = x1;
    sprite.position.y = y1;
};


WorldMap.prototype.drawCircle = function (graphics, lineColor, fillColor, radius, x, y) {
    if (fillColor != null) {
        graphics.beginFill(fillColor, 1);
    }
    graphics.lineStyle(1, lineColor, 1);
    graphics.drawCircle(x, y, radius);
    if (fillColor != null) {
        graphics.endFill();
    }
};

WorldMap.prototype.drawImage = function (graphics, img, x, y, rotation) {};

WorldMap.prototype.drawTriangle = function (graphics, lineColor, fillColor, radius, x, y, rotation) {

    if (fillColor != null) {
        graphics.beginFill(fillColor, 1);
    }
    graphics.lineStyle(1, lineColor, 1);

    let Ax = -0.866 * radius + x;
    let Ay = -0.5 * radius + y;
    let Bx = 0.866 * radius + x;
    let By = -0.5 * radius + y;
    let Cx = 0.0 * radius + x;
    let Cy = 1.0 * radius + y;

    rotatedA = this.rotatePoint(x, y, rotation, Ax, Ay);
    rotatedB = this.rotatePoint(x, y, rotation, Bx, By);
    rotatedC = this.rotatePoint(x, y, rotation, Cx, Cy);

    let A = new PIXI.Point(rotatedA.x, rotatedA.y);
    let B = new PIXI.Point(rotatedB.x, rotatedB.y);
    let C = new PIXI.Point(rotatedC.x, rotatedC.y);

    graphics.drawPolygon(new PIXI.Polygon(A, B, C));
    if (fillColor != null) {
        graphics.endFill();
    }
    graphics.drawPolygon(new PIXI.Polygon(A, B, C));
};

WorldMap.prototype.rotatePoint = function (cx, cy, angle, px, py) {

    var exp = sprintf("((%f - %f) * cos(%frad)) - ((%f - %f) * sin(%frad))", px, cx, angle, py, cy, angle);
    var newx = evaluate(exp);

    exp = sprintf("((%f - %f) * sin(%frad)) + ((%f - %f) * cos(%frad))", px, cx, angle, py, cy, angle);
    var newy = evaluate(exp);

    newx = newx + cx;
    newy = newy + cy;

    return { x: newx, y: newy };
};

module.exports = WorldMap;