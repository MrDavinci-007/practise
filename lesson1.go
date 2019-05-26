//多态
package main

import "fmt"

type people func (string) string

func black(a string)string{
	return fmt.Sprintf("black:hello , %s man",a)
}

func white(a string)string{
	return fmt.Sprintf("white:hello , %s man!!!",a)
}

func person(p string,p2 people) string{
	return p2(p)
}

func main()  {
	fmt.Println(person("block",black))
	fmt.Println(person("white",white))
}
