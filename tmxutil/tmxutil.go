package tmxutil

import "github.com/mewkiz/pkg/imgutil"
import "github.com/mewkiz84/tmx"
import "log"
import "image"

type Map         tmx.Map
type GID         tmx.GID

type SubImager interface {
   image.Image
   SubImage(r image.Rectangle) image.Image
}

// Open reads the provided tmx file and returns a parsed Map, based on the TMX
// file format.
func Open(tmxPath string) (*Map, error) {
   m, err := tmx.Open(tmxPath)
   tm := new(Map)
   tm.Width        = m.Width
   tm.Height       = m.Height
   tm.TileWidth    = m.TileWidth
   tm.TileHeight   = m.TileHeight
   tm.Tilesets     = m.Tilesets
   tm.Layers       = m.Layers
   return tm, err
}

// GetMapRect returns a image.Rectangle with the dimentions of the Map m
func (m *Map) GetMapRect() (rect image.Rectangle) {
   return image.Rect(0, 0, m.Width * m.TileWidth, m.Height * m.TileHeight + m.getMaxTileHeight())
}

// GetTileRect returns the image.Rectangle of the tile on position col, row
func (m *Map) GetTileRect(col, row int) (rect image.Rectangle) {
   blockWidth   := m.TileWidth / 2
   blockHeight  := m.TileHeight / 2
   mapWidth := m.Width * m.TileWidth

   minX := mapWidth/2 - blockWidth - row*blockWidth + col*blockWidth
   minY := row*blockHeight + col*blockHeight

   rect = image.Rect(minX, minY, minX + m.TileWidth, minY + m.TileHeight)

   return rect.Add(image.Pt(0, m.getMaxTileHeight()))
}

// GetRect returns the Rectangle of the  on possition col, row
func (m *Map) GetRect(col, row, GID int, bound image.Rectangle) (rect image.Rectangle) {
   tileRect := m.GetTileRect(col, row)
   offset := m.getTileOffset(GID)
   tileRect.Min.Y += m.TileHeight - bound.Dy()
   tileRect.Max.Y = tileRect.Min.Y + bound.Dy()
   tileRect.Max.X = tileRect.Min.X + bound.Dx()

   return tileRect.Add(image.Pt(offset.X, offset.Y))
}

// GetTileSprite returns a slise of type image.Image containing all tiles from
// all sprites from all layers including the nil tile with GID 0.
func (m *Map) GetTileSprite() (tileSprite []image.Image) {
   tileSprite = make([]image.Image, m.countSpriteTiles())
   k:=1
   for _, tileSet := range m.Tilesets {
      sprite, err := imgutil.ReadFile(tileSet.Image.Source)
      if err != nil {
         log.Fatalln(err)
      }
      spriteImg , ok := sprite.(SubImager)
      if ok != true {
         log.Fatalln("SubImage function not implemented for this image type")
      }
      for y:=0; y<tileSet.Image.Height; y+=tileSet.TileHeight {
         for x:=0; x<tileSet.Image.Width; x+=tileSet.TileWidth {
            rect := image.Rect(x, y, x+tileSet.TileWidth, y+tileSet.TileHeight)
            tileSprite[k] = spriteImg.SubImage(rect)
            k++
         }
      }
   }
   return tileSprite
}

// GlobalTileID returns the GID after clearing the flip flags.
func (gid GID) GlobalTileID() int {
   return int(gid &^ tmx.FlagFlip)
}

// IsDiagonalFlip returns true if the GID is flipped diagonally.
func (gid GID) IsDiagonalFlip() bool {
   if gid&tmx.FlagDiagonalFlip != 0 {
      return true
   }
   return false
}

// IsVerticalFlip returns true if the GID is flipped vertically.
func (gid GID) IsVerticalFlip() bool {
   if gid&tmx.FlagVerticalFlip != 0 {
      return true
   }
   return false
}

// IsHorizontalFlip returns true if the GID is flipped horizontally.
func (gid GID) IsHorizontalFlip() bool {
   if gid&tmx.FlagHorizontalFlip != 0 {
      return true
   }
   return false
}

// IsFlip returns true if the GID is flipped.
func (gid GID) IsFlip() bool {
   if gid&tmx.FlagFlip != 0 {
      return true
   }
   return false
}

// countSpriteTiles returns the total number of tiles on a sprite in a Map including
// the nil tile with GID 0
func (m *Map) countSpriteTiles() (count int) {
   for _, tiles := range m.Tilesets {
      count += tiles.Image.Width/tiles.TileWidth*tiles.Image.Height/tiles.TileHeight
   }
   return count +1
}

func (m *Map) getTileOffset(GID int) (offset tmx.TileOffset) {
   var c, spiteNum int
   for _, sprite := range m.Tilesets {
      c += sprite.Image.Width/sprite.TileWidth*sprite.Image.Height/sprite.TileHeight
      if c > GID {
         break
      }
      spiteNum++
   }
   return tmx.TileOffset(m.Tilesets[spiteNum].TileOffset)
}

func (m *Map) getMaxTileHeight() int {
   var maxHeight int
   for _, sprite := range m.Tilesets {
      if maxHeight < sprite.TileHeight {
         maxHeight = sprite.TileHeight
      }
   }
   return maxHeight
}
