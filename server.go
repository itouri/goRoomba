package main

import (
	"image"
	"net/http"
	"strconv"
	"strings"
	//"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	//"github.com/gocraft/dbr/dialect"

	"image/jpeg"
	"os"
	"os/exec"
)

type (
	userinfo struct {
		ID         int    `db:"id"`
		Email      string `db:"email"`
		First_name string `db:"first_name"`
		Last_name  string `db:"last_name"`
		Image      byte   `db:image`
	}

	userinfoJSON struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Firstname string `json:"firstName"`
		Lastname  string `json:"lastName"`
		Image     byte   `db:image`
	}

	responseData struct {
		Users []userinfo `json:"users"`
	}

	positionStr struct {
		PosY int `json:"posY"`
		PosY int `json:"posY"`
	}

	robotsJSON struct {
		ID       int    `json:"id"`
		Position [2]int `json:"position"`
	}

	robotsArrayJSON struct {
		Robots []robotsJSON `json:"robots"`
	}

	lostPropertiesJSON struct {
		ID    int    `json:"id"`
		Image string `json:"image"`
	}

	lostPropertiesArrayJSON struct {
		LostProperties []lostPropertiesJSON
	}

	regularContactJSON struct {
		ID       int    `json:"id"`
		Position [2]int `json:"position"`
		Image    string `json:"image"`
	}

	searchRangeJSON struct {
		ID     int `json:"id"`
		StartY int `json:"startY"`
		EndY   int `json:"endY"`
	}
)

var (
	tablename = "userinfo"
	seq       = 1
	conn, _   = dbr.Open("mysql", "root:@tcp(127.0.0.1:3306)/test", nil)
	sess      = conn.NewSession(nil)
)

var robotLocations map[int][2]int
var lostProperties map[int]string
var connctedRobotNum int
var currentID int

//----------
// Handlers
//----------

func putAny(c echo.Context) error {
	u := new(userinfoJSON)
	if err := c.Bind(u); err != nil {
		return err
	}

	photo, _ := c.FormFile("photo")
	src, _ := photo.Open()
	img, _, _ := image.Decode(src)

	file, err := os.Create("recv.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	//file.Write(img)

	outFile, _ := os.Create("save.jpg")
	option := &jpeg.Options{Quality: 100}
	jpeg.Encode(outFile, img, option)

	sess.InsertInto(tablename).Values(u).Exec()
	//sess.InsertInto(tablename).Columns("id","email","first_name","last_name").Values(u.ID,u.Email,u.Firstname,u.Lastname).Exec()

	return c.NoContent(http.StatusOK)
}

func insertUser(c echo.Context) error {
	u := new(userinfoJSON)
	if err := c.Bind(u); err != nil {
		return err
	}

	sess.InsertInto(tablename).Columns("id", "email", "first_name", "last_name", "image").Values(u.ID, u.Email, u.Firstname, u.Lastname, u.Image).Exec()

	return c.NoContent(http.StatusOK)
}

func selectUsers(c echo.Context) error {
	var u []userinfo

	sess.Select("*").From(tablename).Load(&u)
	response := new(responseData)
	response.Users = u
	return c.JSON(http.StatusOK, response)
}
func selectUser(c echo.Context) error {
	var m userinfo
	id := c.Param("id")
	sess.Select("*").From(tablename).Where("id = ?", id).Load(&m)
	//idはPrimary Key属性のため重複はありえない。そのため一件のみ取得できる。そのため配列である必要はない
	return c.JSON(http.StatusOK, m)

}

func updateUser(c echo.Context) error {
	u := new(userinfoJSON)
	if err := c.Bind(u); err != nil {
		return err
	}

	attrsMap := map[string]interface{}{"id": u.ID, "email": u.Email, "first_name": u.Firstname, "last_name": u.Lastname, "image": u.Image}
	sess.Update(tablename).SetMap(attrsMap).Where("id = ?", u.ID).Exec()
	return c.NoContent(http.StatusOK)
}

func deleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	sess.DeleteFrom(tablename).
		Where("id = ?", id).
		Exec()

	return c.NoContent(http.StatusNoContent)
}

func getRobots(c echo.Context) error {
	response := new(robotsArrayJSON)
	for key, value := range robotLocations {
		robot := robotsJSON{
			ID:       key,
			Position: value,
		}
		response.Robots = append(response.Robots, robot)
	}
	return c.JSON(http.StatusOK, response)
}

func getLostProperties(c echo.Context) error {
	response := new(lostPropertiesArrayJSON)
	for key, value := range lostProperties {
		lp := lostPropertiesJSON{
			ID:    key,
			Image: value,
		}
		response.LostProperties = append(response.LostProperties, lp)
	}
	return c.JSON(http.StatusOK, response)
}

func postRegularContact(c echo.Context) error {
	rc := new(regularContactJSON)
	if err := c.Bind(rc); err != nil {
		return err
	}
	robotLocations[rc.ID] = rc.Position
	response := searchRangeJSON{
		ID:     -1,
		StartY: 10,
		EndY:   10,
	}
	if false {
		return c.JSON(http.StatusOK, response)
	} else {
		return c.NoContent(http.StatusNoContent)
	}
}

func postFirstContact(c echo.Context) error {
	currentID++
	connctedRobotNum++
	rc := new(regularContactJSON)
	if err := c.Bind(rc); err != nil {
		return err
	}
	robotLocations[rc.ID] = rc.Position
	res := []searchRangeJSON{}
	out, err := exec.Command("./separete", "Image_0259ce6.bmp", strconv.Itoa(connctedRobotNum)).Output()
	if err != nil {
		return err
	}
	outStr := string(out[:])
	posAry := strings.Split(outStr, ",")
	startY := 0
	for _, posY := range posAry {
		endY, _ := strconv.Atoi(posY)
		addRes := searchRangeJSON{
			ID:     currentID,
			StartY: startY,
			EndY:   endY,
		}
		startY = endY + 1
		res = append(res, addRes)
	}
	return c.JSON(http.StatusOK, res)
}

func main() {
	e := echo.New()

	robotLocations = make(map[int][2]int)
	lostProperties = make(map[int]string)
	connctedRobotNum = 0
	currentID = 0

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes

	e.POST("/users/", insertUser)
	e.GET("/user/:id", selectUser)
	e.GET("/users/", selectUsers)
	e.PUT("/users/", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// for Administrator
	e.GET("/robots/", getRobots)
	e.GET("/lost-properties/", getLostProperties)

	// from Roomba
	e.POST("/regular-contact/", postRegularContact)
	e.GET("/first-contact/", postFirstContact)

	e.POST("/", putAny)

	// Start server
	//e.Run(standard.New(":1323"))
	e.Start(":1323")
}
