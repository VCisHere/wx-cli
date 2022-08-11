package client

import (
	"fmt"
)

type Friend struct {
	*User
}

// implement fmt.Stringer
func (f Friend) String() string {
	return fmt.Sprintf("<Friend:%s>", f.NickName)
}

type Friends []*Friend

// Count 获取好友的数量
func (f Friends) Count() int {
	return len(f)
}

// First 获取第一个好友
func (f Friends) First() *Friend {
	if f.Count() > 0 {
		return f[0]
	}
	return nil
}

// Last 获取最后一个好友
func (f Friends) Last() *Friend {
	if f.Count() > 0 {
		return f[f.Count()-1]
	}
	return nil
}

// SearchByUserName 根据用户名查找好友
func (f Friends) SearchByUserName(limit int, username string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.UserName == username })
}

// SearchByNickName 根据昵称查找好友
func (f Friends) SearchByNickName(limit int, nickName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.NickName == nickName })
}

// SearchByRemarkName 根据备注查找好友
func (f Friends) SearchByRemarkName(limit int, remarkName string) (results Friends) {
	return f.Search(limit, func(friend *Friend) bool { return friend.User.RemarkName == remarkName })
}

// Search 根据自定义条件查找好友
func (f Friends) Search(limit int, condFuncList ...func(friend *Friend) bool) (results Friends) {
	if condFuncList == nil {
		return f
	}
	if limit <= 0 {
		limit = f.Count()
	}
	for _, member := range f {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

type Group struct {
	*User
}

// Members 获取所有的群成员
func (g *Group) Members(self *Self) (Members, error) {
	if err := g.Detail(self); err != nil {
		return nil, err
	}
	g.MemberList.init()
	return g.MemberList, nil
}

type Groups []*Group

// Count 获取群组数量
func (g Groups) Count() int {
	return len(g)
}

// First 获取第一个群组
func (g Groups) First() *Group {
	if g.Count() > 0 {
		return g[0]
	}
	return nil
}

// Last 获取最后一个群组
func (g Groups) Last() *Group {
	if g.Count() > 0 {
		return g[g.Count()-1]
	}
	return nil
}

// SearchByUserName 根据用户名查找群组
func (g Groups) SearchByUserName(limit int, username string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.UserName == username })
}

// SearchByNickName 根据昵称查找群组
func (g Groups) SearchByNickName(limit int, nickName string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.NickName == nickName })
}

// SearchByRemarkName 根据备注查找群组
func (g Groups) SearchByRemarkName(limit int, remarkName string) (results Groups) {
	return g.Search(limit, func(group *Group) bool { return group.RemarkName == remarkName })
}

// Search 根据自定义条件查找群组
func (g Groups) Search(limit int, condFuncList ...func(group *Group) bool) (results Groups) {
	if condFuncList == nil {
		return g
	}
	if limit <= 0 {
		limit = g.Count()
	}
	for _, member := range g {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

// Mp 公众号对象
type Mp struct {
	*User
}

func (m Mp) String() string {
	return fmt.Sprintf("<Mp:%s>", m.NickName)
}

// Mps 公众号组对象
type Mps []*Mp

// Count 数量统计
func (m Mps) Count() int {
	return len(m)
}

// First 获取第一个
func (m Mps) First() *Mp {
	if m.Count() > 0 {
		return m[0]
	}
	return nil
}

// Last 获取最后一个
func (m Mps) Last() *Mp {
	if m.Count() > 0 {
		return m[m.Count()-1]
	}
	return nil
}

// Search 根据自定义条件查找
func (m Mps) Search(limit int, condFuncList ...func(group *Mp) bool) (results Mps) {
	if condFuncList == nil {
		return m
	}
	if limit <= 0 {
		limit = m.Count()
	}
	for _, member := range m {
		if results.Count() == limit {
			break
		}
		var passCount int
		for _, condFunc := range condFuncList {
			if condFunc(member) {
				passCount++
			}
		}
		if passCount == len(condFuncList) {
			results = append(results, member)
		}
	}
	return
}

// SearchByUserName 根据用户名查找
func (m Mps) SearchByUserName(limit int, userName string) (results Mps) {
	return m.Search(limit, func(group *Mp) bool { return group.UserName == userName })
}

// SearchByNickName 根据昵称查找
func (m Mps) SearchByNickName(limit int, nickName string) (results Mps) {
	return m.Search(limit, func(group *Mp) bool { return group.NickName == nickName })
}

// GetByUsername 根据username查询一个Friend
func (f Friends) GetByUsername(username string) *Friend {
	return f.SearchByUserName(1, username).First()
}

// GetByRemarkName 根据remarkName查询一个Friend
func (f Friends) GetByRemarkName(remarkName string) *Friend {
	return f.SearchByRemarkName(1, remarkName).First()
}

// GetByNickName 根据nickname查询一个Friend
func (f Friends) GetByNickName(nickname string) *Friend {
	return f.SearchByNickName(1, nickname).First()
}

// GetByUsername 根据username查询一个Group
func (g Groups) GetByUsername(username string) *Group {
	return g.SearchByUserName(1, username).First()
}

// GetByRemarkName 根据remarkName查询一个Group
func (g Groups) GetByRemarkName(remarkName string) *Group {
	return g.SearchByRemarkName(1, remarkName).First()
}

// GetByNickName 根据nickname查询一个Group
func (g Groups) GetByNickName(nickname string) *Group {
	return g.SearchByNickName(1, nickname).First()
}

// GetByNickName 根据nickname查询一个Mp
func (m Mps) GetByNickName(nickname string) *Mp {
	return m.SearchByNickName(1, nickname).First()
}

// GetByUserName 根据username查询一个Mp
func (m Mps) GetByUserName(username string) *Mp {
	return m.SearchByUserName(1, username).First()
}
