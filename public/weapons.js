// Weapons
const WEAPONS = {};
WEAPONS[WEAPONS["WEAPON_Grenade"] = 0] = "WEAPON_Grenade";

/**
 * @class Weapon
 */
class Weapon {
    player_index;
    constructor(player_index) {
        if(this.constructor == Weapon) {
            throw new Error("Abstract classes can't be instantiated.");
        }
        this.player_index = player_index;
    }

    onWeaponPressedMSG(msg,player_index) {}
    onWeaponUpdatedMSG(msg,player_index) {}
    onWeaponReleasedMSG(msg,player_index) {}
    render(ctx){}

    static getWeaponObject(id,index) {
    
        switch (WEAPONS[id]) {
            case "WEAPON_Grenade":
                return new Grenade(index);
            default:
                return null;
        }
    }

    static decodeWeaponPressedMSG(msg) {
        const player_id = getInt16(new DataView(msg, 2), { i: 0 });
        let player_index = game.state.players.findIndex(player => player.user.id == player_id);
        
        return game.state.players[player_index].weapon.onWeaponPressedMSG(msg);
    }

    static decodeWeaponUpdatedMSG(msg) {
        const player_id = getInt16(new DataView(msg, 2), { i: 0 });

        let player_index = game.state.players.findIndex(player => player.user.id == player_id);

        return game.state.players[player_index].weapon.onWeaponUpdatedMSG(msg);
    }

    static decodeWeaponReleasedMSG(msg) {
        const player_id = getInt16(new DataView(msg, 2), { i: 0 });

        let player_index = game.state.players.findIndex(player => player.user.id == player_id);

        return game.state.players[player_index].weapon.onWeaponReleasedMSG(msg);
    }

}

class Grenade extends Weapon {
    isAiming = false;
    isShooting=false;
    seta=null;
    target=null;
    constructor(player_index) {
        super(player_index);
        
    }

    onWeaponPressedMSG(msg) {
        this.isAiming=true;
        return this.onWeaponUpdatedMSG(msg);
    }

    onWeaponUpdatedMSG(msg) {
        const view = new DataView(msg, 4);
        const state = { i: 0 };
        const data = {};
        
        // check the length of the message
        let msg_len = getUint8(view, state);
        if(msg_len != 8){
            throw new Error("Invalid message length "+msg_len);
        }

        this.seta = getFloat64(view, state);
        
        return data;
    }

    onWeaponReleasedMSG(msg) {
        const view = new DataView(msg, 4);
        const state = { i: 0 };
        const data = {};
        
        // check the length of the message
        let msg_len = getUint8(view, state);
        if(msg_len != 16){
            throw new Error("Invalid message length "+msg_len);
        }

        data.x = getFloat64(view, state);
        data.y = getFloat64(view, state);

        //get the tiles state by getting the teamId and add 1 to it to get the tile code.
        let tiles_state=game.state.players[this.player_index].team + 1

        this.target = data;
        this.isShooting=true;
        this.isAiming=false;
        this.seta=null;

        setTimeout(() => {
            this.onWeaponReleased(data.x,data.y, tiles_state);
        }, 100);

        setTimeout(() => {
            this.isShooting=false;
            this.target=null;
        }, 600);
        
        return data;
    }

    onWeaponReleased(x,y,state){
        x=Math.floor(x);
        y=Math.floor(y);

        for(let i=x-1;i<x+2;i++){
            for(let j=y-1;j<y+2;j++){
                let tile_index=(j * game.map.width + i)
                
                //if the tile is a wall continue
                if(game.map.tiles[tile_index]==WallTile)continue

                if(i>=0 && j>=0 && i<game.map.width && j<game.map.height){
                    game.map.tiles[tile_index] = state;
                }
            }
        }
    }

    render(ctx){
        
        if(!this.isAiming && !this.isShooting){
            return;
        }

        let player=game.state.players[this.player_index];

        if(this.isAiming){

            let rect=document.getElementById("canvas").getBoundingClientRect()
            let cellInPixels= ((1600/rect.width)+(900/rect.height))/2
            
            let y = (player.y + 0.5) - (0.4 * cellInPixels * Math.sin(this.seta)) 
		    let x = (player.x + 0.5) - (0.4 * cellInPixels * Math.cos(this.seta))
            
            ctx.lineWidth = 8; 
            ctx.beginPath();
            ctx.strokeStyle = "green";
            ctx.moveTo(...mapCellToCanvasPixel(player.x + 0.5, player.y + 0.5));
            ctx.lineTo(...mapCellToCanvasPixel(x, y));
            ctx.stroke();  

        }else if(this.isShooting){
            ctx.lineWidth = 5; 
            ctx.beginPath();
            ctx.strokeStyle = "red";
            ctx.moveTo(...mapCellToCanvasPixel(player.x + 0.5, player.y + 0.5));
            ctx.lineTo(...mapCellToCanvasPixel(this.target.x, this.target.y));
            ctx.stroke();  
        }
    }

    /** @todo: use this class to handle the own player weapon actions */
}