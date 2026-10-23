package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var filePath = "index.html"

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n========= Minecraft 数据库后台管理 =========")
		fmt.Println("1. 修改公告")
		fmt.Println("2. 添加帖子 (支持自定义头像、封面、富文本)")
		fmt.Println("3. 删除帖子")
		fmt.Println("4. 添加轮播广告 (支持文字与链接)")
		fmt.Println("5. 添加展示位广告 (指定位置)")
		fmt.Println("6. 添加商品/服务卡片 (支持勾选与备注)")
		fmt.Println("7. 删除商品/服务卡片")
		fmt.Println("8. 删除轮播广告")
		fmt.Println("9. 添加壁纸")
		fmt.Println("10. 删除壁纸")
		fmt.Println("0. 退出")
		fmt.Print("请选择操作: ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			modifyAnnouncement(reader)
		case "2":
			addPost(reader)
		case "3":
			deletePost(reader)
		case "4":
			addCarouselAd(reader)
		case "5":
			addSlotAd(reader)
		case "6":
			addProduct(reader)
		case "7":
			deleteProduct(reader)
		case "8":
			deleteCarouselAd(reader)
		case "9":
			addWallpaper(reader)
		case "10":
			deleteWallpaper(reader)
		case "0":
			fmt.Println("退出工具。")
			return
		default:
			fmt.Println("输入无效，请重试。")
		}
	}
}

// 读取并解析 HTML (仅用于 DOM 操作)
func loadDoc() *goquery.Document {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("❌ 找不到 index.html，请确认文件在当前目录！")
		os.Exit(1)
	}
	defer file.Close()
	
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		fmt.Println("❌ 解析 HTML 失败:", err)
		os.Exit(1)
	}
	return doc
}

// 备份并保存 HTML
func saveDoc(doc *goquery.Document) {
	backupName := fmt.Sprintf("index_backup_%s.html", time.Now().Format("20060102_150405"))
	oldData, err := os.ReadFile(filePath)
	if err == nil {
		os.WriteFile(backupName, oldData, 0644)
		fmt.Printf("✅ 已备份旧文件为: %s\n", backupName)
	}

	html, err := doc.Html()
	if err != nil {
		fmt.Println("❌ 生成 HTML 失败:", err)
		return
	}

	err = os.WriteFile(filePath, []byte(html), 0644)
	if err != nil {
		fmt.Println("❌ 保存文件失败:", err)
		return
	}
	fmt.Println("✅ 修改成功！新的 index.html 已生成。")
}

// 工具函数：获取用户输入
func getInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// 1. 修改公告
func modifyAnnouncement(reader *bufio.Reader) {
	doc := loadDoc()
	newText := getInput(reader, "请输入新的公告内容: ")

	doc.Find(".announce span").Last().SetText(newText)
	doc.Find(".rail-panel").Each(func(i int, s *goquery.Selection) {
		if strings.TrimSpace(s.Find("h3").Text()) == "公告" {
			s.Find("p").SetText(newText)
		}
	})
	saveDoc(doc)
}

// 2. 添加帖子
func addPost(reader *bufio.Reader) {
	doc := loadDoc()
	
	title := getInput(reader, "标题: ")
	content := getInput(reader, "摘要(用于搜索): ")
	tags := getInput(reader, "标签(如: 官方,安卓,Java版 英文逗号隔开): ")
	author := getInput(reader, "作者名: ")
	
	avatarUrl := getInput(reader, "头像图片链接 (没有则直接回车，使用默认): ")
	
	isOfficial := getInput(reader, "是否官方(输入 1 是，0 否): ")
	
	fmt.Println("【帖子类型】")
	postType := getInput(reader, "1: 本地详情页, 2: 直接跳转外链: ")
	var externalUrl string
	if postType == "2" {
		externalUrl = getInput(reader, "请输入跳转外链地址 (如 https://...): ")
	}

	fmt.Println("【封面设置】")
	coverType := getInput(reader, "1: 图片链接, 2: 纯文字, 直接回车: 使用默认: ")
	var coverHTML string
	if coverType == "1" {
		imgUrl := getInput(reader, "请输入图片链接: ")
		coverHTML = fmt.Sprintf(`<div class="post-cover" style="background-image: url('%s'); background-size: cover; background-position: center;"></div>`, imgUrl)
	} else if coverType == "2" {
		coverText := getInput(reader, "请输入封面文字 (如: ZL2官方下载): ")
		coverHTML = fmt.Sprintf(`<div class="post-cover cat-download" style="font-size:1.1rem;font-weight:700;letter-spacing:1px;opacity:1;">%s</div>`, coverText)
	} else {
		coverHTML = `<div class="post-cover cat-video">▶</div>`
	}

	pos := getInput(reader, "插入位置 (1: 置顶, 2: 放在最下方): ")
	
	articleID := fmt.Sprintf("post-%d", time.Now().Unix())
	currentDate := time.Now().Format("2006.01.02")

	badge := ""
	if isOfficial == "1" {
		badge = `<span class="official-badge">官方</span>`
	}

	tagSpans := ""
	for _, t := range strings.Split(tags, ",") {
		tagSpans += fmt.Sprintf(`<span class="cat-tag">%s</span>`, strings.TrimSpace(t))
	}

	var avatarHTML string
	if avatarUrl != "" {
		avatarHTML = fmt.Sprintf(`<img class="avatar" src="%s" alt="avatar">`, avatarUrl)
	} else {
		avatarHTML = `<span class="avatar"></span>`
	}

	var cardOnclick, titleOnclick string
	if postType == "2" {
		cardOnclick = fmt.Sprintf(`window.open('%s', '_blank')`, externalUrl)
		titleOnclick = `event.stopPropagation()`
	} else {
		cardOnclick = fmt.Sprintf(`go('%s')`, articleID)
		titleOnclick = fmt.Sprintf(`event.stopPropagation(); go('%s')`, articleID)
	}

	cardHTML := fmt.Sprintf(`
	<article class="glass-soft post-card" data-title="%s" data-content="%s" data-tags="%s" onclick="%s">
		%s
		<div class="post-body">
			<div class="post-byline">
				%s %s %s · %s
			</div>
			<a class="post-title" href="javascript:void(0)" onclick="%s">%s</a>
			<p class="post-excerpt">%s</p>
			<div style="display:flex;gap:6px;flex-wrap:wrap">%s</div>
		</div>
	</article>`, title, content, tags, cardOnclick, coverHTML, avatarHTML, author, badge, currentDate, titleOnclick, title, content, tagSpans)

	if pos == "1" {
		doc.Find("#postList").PrependHtml(cardHTML)
	} else {
		doc.Find("#postList").AppendHtml(cardHTML)
	}

	if postType == "1" {
		fmt.Println("【正文输入】")
		fmt.Println("支持 HTML 标签 (如 <h2>大标题</h2>, <b>粗体</b>, <a href=\"...\">链接</a>)")
		fmt.Println("多行输入。输入完内容后，请单独起一行输入 END 并回车结束：")
		
		var bodyLines []string
		for {
			line, _ := reader.ReadString('\n')
			line = strings.TrimRight(line, "\r\n")
			if line == "END" {
				break
			}
			bodyLines = append(bodyLines, line)
		}
		body := strings.Join(bodyLines, "\n")

		detailHTML := fmt.Sprintf(`
	<section class="page" id="page-%s">
		<div class="glass doc" style="max-width: 800px; margin: 0 auto;">
			<div class="article-header">
				<h1>%s</h1>
				<div class="post-byline" style="justify-content: center;">
					%s %s %s · %s
				</div>
				<div style="display:flex;gap:6px;justify-content:center;margin-top:12px;flex-wrap:wrap;">%s</div>
			</div>
			<div class="article-content">
				%s
				<div style="margin-top:28px; padding-top:20px; border-top:1px solid var(--line); display:flex; gap:12px; flex-wrap:wrap; justify-content:center;">
					<button class="glass-pill btn btn-ghost" onclick="go('home')">返回首页</button>
				</div>
			</div>
		</div>
	</section>`, articleID, title, avatarHTML, author, badge, currentDate, tagSpans, body)
		
		doc.Find("main").AppendHtml(detailHTML)
	}

	saveDoc(doc)
}

// 3. 删除帖子
func deletePost(reader *bufio.Reader) {
	doc := loadDoc()
	title := getInput(reader, "请输入要删除的帖子【标题】(必须完全一致): ")
	
	var targetID string
	doc.Find("article.post-card").Each(func(i int, s *goquery.Selection) {
		if strings.TrimSpace(s.AttrOr("data-title", "")) == title {
			onclickAttr := s.AttrOr("onclick", "")
			if strings.Contains(onclickAttr, "go('") {
				start := strings.Index(onclickAttr, "go('") + 4
				end := strings.Index(onclickAttr[start:], "')")
				if end != -1 {
					targetID = onclickAttr[start : start+end]
				}
			}
			s.Remove()
		}
	})

	if targetID != "" {
		fmt.Printf("✅ 首页卡片已删除，匹配到的内部ID为: %s\n", targetID)
		doc.Find(fmt.Sprintf("#page-%s", targetID)).Remove()
		fmt.Println("✅ 对应详情页也已同步删除。")
	} else {
		fmt.Println("⚠️ 未找到对应标题的帖子卡片，或者该帖子是外链帖(无独立详情页)。")
	}

	saveDoc(doc)
}

// 4. 添加轮播广告
func addCarouselAd(reader *bufio.Reader) {
	doc := loadDoc()
	fmt.Println("【添加轮播广告】")
	imgUrl := getInput(reader, "请输入背景图片链接 (填 无 则使用默认渐变): ")
	text := getInput(reader, "请输入显示的文字 (没有则直接回车): ")
	link := getInput(reader, "请输入点击跳转的链接 (没有则直接回车): ")
	
	var slideHTML string
	if imgUrl != "无" && imgUrl != "" {
		slideHTML = fmt.Sprintf(`<div class="carousel-slide" style="background-image: url('%s');">`, imgUrl)
	} else {
		slideHTML = `<div class="carousel-slide" style="background:linear-gradient(135deg,#667eea,#764ba2);">`
	}

	if text != "" {
		if link != "" {
			slideHTML += fmt.Sprintf(`<a href="%s" target="_blank" style="display:flex; width:100%%; height:100%%; align-items:center; justify-content:center; color:#fff; font-size:1.5rem; font-weight:bold; text-decoration:none; text-shadow:0 2px 8px rgba(0,0,0,0.8);">%s</a>`, link, text)
		} else {
			slideHTML += fmt.Sprintf(`<div style="display:flex; width:100%%; height:100%%; align-items:center; justify-content:center; color:#fff; font-size:1.5rem; font-weight:bold; text-shadow:0 2px 8px rgba(0,0,0,0.8);">%s</div>`, text)
		}
	} else if link != "" {
		slideHTML += fmt.Sprintf(`<a href="%s" target="_blank" style="display:block; width:100%%; height:100%%;"></a>`, link)
	}

	slideHTML += `</div>`
	
	doc.Find("#adCarousel").PrependHtml(slideHTML)
	saveDoc(doc)
}

// 5. 添加展示位广告
func addSlotAd(reader *bufio.Reader) {
	doc := loadDoc()
	link := getInput(reader, "跳转链接 (没有则直接回车): ")
	imgUrl := getInput(reader, "图片链接(填 无 则只显示文字): ")
	text := getInput(reader, "显示文字: ")

	fmt.Println("当前已有位置: 1(QQ群), 2, 3, 4, 5, 6 ...")
	targetStr := getInput(reader, "你想把广告放在第几个格子里？(输入数字，如 3): ")
	targetIdx, err := strconv.Atoi(targetStr)
	if err != nil || targetIdx < 1 {
		fmt.Println("❌ 输入的数字无效，默认放到最后。")
		targetIdx = 999
	}
	
	var innerHTML string
	if imgUrl != "无" && imgUrl != "" {
		innerHTML = fmt.Sprintf(`<img src="%s" style="position:absolute; inset:0; width:100%%; height:100%%; object-fit:cover; z-index:0;">
		<span style="position:relative; z-index:1; color:var(--text); font-weight:bold; text-shadow:0 1px 3px rgba(0,0,0,0.3); font-size:1rem; letter-spacing:1px;">%s</span>`, imgUrl, text)
	} else {
		innerHTML = fmt.Sprintf(`<span class="ad-num">新</span>%s`, text)
	}

	slotHTML := fmt.Sprintf(`<a href="%s" target="_blank" class="glass-soft ad-slot" style="padding:0; overflow:hidden; border:none; position:relative;">%s</a>`, link, innerHTML)

	slots := doc.Find(".ad-quicklinks .ad-slot")
	currentCount := slots.Length()

	for i := currentCount; i < targetIdx-1; i++ {
		num := i + 2
		doc.Find(".ad-quicklinks").AppendHtml(fmt.Sprintf(`<div class="glass-soft ad-slot"><span class="ad-num">%d</span>展示位 %d</div>`, num, num))
	}

	if targetIdx <= currentCount {
		slots.Eq(targetIdx - 1).ReplaceWithHtml(slotHTML)
		fmt.Printf("✅ 已替换第 %d 个格子。\n", targetIdx)
	} else {
		doc.Find(".ad-quicklinks").AppendHtml(slotHTML)
		fmt.Println("✅ 已添加到最末尾。")
	}

	saveDoc(doc)
}

// 6. 添加商品
func addProduct(reader *bufio.Reader) {
	doc := loadDoc()
	title := getInput(reader, "商品标题: ")
	desc := getInput(reader, "商品描述: ")
	
	fmt.Println("【商品勾选列表】 (输入你想打勾的内容，可多行输入，输入 END 结束。直接打 END 跳过)")
	var listItems []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line == "END" {
			break
		}
		if line != "" {
			listItems = append(listItems, line)
		}
	}
	
	var listHTML string
	if len(listItems) > 0 {
		listHTML = `<ul class="checklist">`
		for _, item := range listItems {
			listHTML += fmt.Sprintf(`<li>%s</li>`, item)
		}
		listHTML += `</ul>`
	}

	price := getInput(reader, "价格 (如: 8.9): ")
	btnText := getInput(reader, "按钮文字 (如: 联系我下单): ")
	remark := getInput(reader, "底部备注 (没有则直接回车): ")
	
	var remarkHTML string
	if remark != "" {
		remarkHTML = fmt.Sprintf(`<p style="font-size:.78rem;color:var(--muted);margin-bottom:0">%s</p>`, remark)
	}

	productHTML := fmt.Sprintf(`
	<div class="glass service-card">
		<h2>%s</h2>
		<p>%s</p>
		%s
		<div class="price-row">
			<div class="price">¥%s <span>/ 次</span></div>
			<button class="glass-pill btn btn-primary" onclick="go('order')">%s</button>
		</div>
		%s
	</div>`, title, desc, listHTML, price, btnText, remarkHTML)
	
	doc.Find("#page-service").AppendHtml(productHTML)
	saveDoc(doc)
}

// 7. 删除商品
func deleteProduct(reader *bufio.Reader) {
	doc := loadDoc()
	title := getInput(reader, "请输入要删除的商品标题 (如: 正版游戏购买指导): ")
	doc.Find(".service-card").Each(func(i int, s *goquery.Selection) {
		h2 := s.Find("h2").Text()
		if strings.TrimSpace(h2) == title {
			s.Remove()
			fmt.Println("✅ 已找到并删除该商品。")
		}
	})
	saveDoc(doc)
}

// 8. 删除轮播广告
func deleteCarouselAd(reader *bufio.Reader) {
	doc := loadDoc()
	fmt.Println("【当前轮播广告列表】")
	slides := doc.Find(".carousel-slide")
	if slides.Length() == 0 {
		fmt.Println("❌ 当前没有任何轮播广告。")
		return
	}

	slides.Each(func(i int, s *goquery.Selection) {
		style, _ := s.Attr("style")
		text := s.Text()
		fmt.Printf("[%d] 样式: %s | 文字: %s\n", i+1, style, text)
	})

	idxStr := getInput(reader, "请输入要删除的轮播广告序号 (从1开始): ")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 1 || idx > slides.Length() {
		fmt.Println("❌ 输入的数字无效。")
		return
	}

	slides.Eq(idx - 1).Remove()
	fmt.Printf("✅ 第 %d 个轮播广告已删除。\n", idx)

	saveDoc(doc)
}

// 9. 添加壁纸
func addWallpaper(reader *bufio.Reader) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ 读取文件失败！")
		return
	}
	content := string(data)
	newUrl := getInput(reader, "请输入新壁纸的图片链接: ")
	
	re := regexp.MustCompile(`(?s)const backgroundList = \[(.*?)\];`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		inner := strings.TrimSpace(matches[1])
		if inner != "" {
			inner += ",\n    '" + newUrl + "'"
		} else {
			inner = "'" + newUrl + "'"
		}
		newBlock := "const backgroundList = [\n    " + inner + "\n  ];"
		content = re.ReplaceAllString(content, newBlock)
		
		backupName := fmt.Sprintf("index_backup_%s.html", time.Now().Format("20060102_150405"))
		os.WriteFile(backupName, data, 0644)
		os.WriteFile(filePath, []byte(content), 0644)
		fmt.Println("✅ 壁纸添加成功！")
	} else {
		fmt.Println("❌ 未找到壁纸配置块，无法添加。")
	}
}

// 10. 删除壁纸 (终极修复版：彻底忽略默认壁纸，强制保留)
func deleteWallpaper(reader *bufio.Reader) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ 读取文件失败！")
		return
	}
	content := string(data)
	
	re := regexp.MustCompile(`(?s)const backgroundList = \[(.*?)\];`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		inner := matches[1]
		
		// 使用正则精准提取所有的 http/https 链接，过滤掉空字符串 ''
		urlRe := regexp.MustCompile(`'(https?://[^']+)'`)
		rawUrls := urlRe.FindAllStringSubmatch(inner, -1)
		
		var realUrls []string
		for _, u := range rawUrls {
			realUrls = append(realUrls, u[1]) // u[1] 是去掉引号的纯链接
		}
		
		if len(realUrls) == 0 {
			fmt.Println("❌ 当前没有自定义壁纸可以删除。")
			return
		}
		
		fmt.Println("【当前可删除的壁纸列表】")
		for i, u := range realUrls {
			fmt.Printf("[%d] %s\n", i+1, u)
		}
		
		idxStr := getInput(reader, "请输入要删除的壁纸序号: ")
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 1 || idx > len(realUrls) {
			fmt.Println("❌ 输入错误。")
			return
		}
		
		// 重新拼接数组，永远确保第一个是默认壁纸 ''
		var newItems []string
		newItems = append(newItems, "''") 
		for i, u := range realUrls {
			if i+1 != idx {
				newItems = append(newItems, "'"+u+"'")
			}
		}
		
		newBlock := "const backgroundList = [\n    " + strings.Join(newItems, ",\n    ") + "\n  ];"
		content = re.ReplaceAllString(content, newBlock)
		
		backupName := fmt.Sprintf("index_backup_%s.html", time.Now().Format("20060102_150405"))
		os.WriteFile(backupName, data, 0644)
		os.WriteFile(filePath, []byte(content), 0644)
		fmt.Println("✅ 壁纸删除成功！默认壁纸已保留。")
	} else {
		fmt.Println("❌ 未找到壁纸配置块。")
	}
}