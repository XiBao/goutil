package ecom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/XiBao/goutil"
)

type Platform int

const (
	UNKNOWN_PLATFORM Platform = iota
	TAOBAO
	JD
	PDD
	WECHAT
	MEITUAN
)

func linkCacheKey(link string) string {
	return goutil.Md5(goutil.StringsJoin("external_url:", link))
}

type ICache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, exp int64) error
}

var cache ICache

func SetCache(v ICache) {
	cache = v
}

func Cache() ICache {
	return cache
}

func ExtractLink(ctx context.Context, link string, cacheExp int64) (string, error) {
	parsedLink, err := url.ParseRequestURI(link)
	if err != nil {
		return "", errors.Join(errors.New("解析链接错误"), err)
	}
	if strings.HasSuffix(parsedLink.Host, ".xibao100.com") {
		return link, nil
	}
	suffixDomains := []string{".taobao.com", ".tmall.com", ".tmall.hk", ".jd.com", ".jd.hk", ".yiyaojd.com", ".tb.cn", ".pinduoduo.com", ".yangkeduo.com", ".duanqu.com", ".1688.com", ".meituan.com"}
	for _, suffix := range suffixDomains {
		if strings.HasSuffix(parsedLink.Host, suffix) {
			return link, nil
		}
	}
	if cacheExp > 0 && Cache() != nil {
		if oriLink, err := Cache().Get(ctx, linkCacheKey(link)); err == nil {
			return oriLink, nil
		}
	}
	if strings.HasPrefix(path.Clean(parsedLink.Path), "/landing/") {
		resp, err := http.DefaultClient.Get(link)
		if err != nil {
			return "", errors.Join(errors.New("下载链接内容失败"), err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Join(errors.New("读取链接内容失败"), err)
		}
		re := regexp.MustCompile(`<a\s+href="(https://.+?)"`)
		matches := re.FindAllSubmatch(body, -1)
		if len(matches) == 0 {
			re = regexp.MustCompile(`,link:"(https.+?)",`)
			matches := re.FindAllSubmatch(body, -1)
			if len(matches) == 0 || len(matches[0]) < 2 {
				re = regexp.MustCompile(`openUrl\('(https.+?)'\)`)
				matches = re.FindAllSubmatch(body, -1)
				if len(matches) == 0 || len(matches[0]) < 2 {
					return "", errors.Join(errors.New("提取二跳链接失败"), nil)
				}
			}
			matched := string(matches[0][1])
			ret := strings.ReplaceAll(strings.ReplaceAll(matched, `\/`, "/"), `\u0026`, "&")
			if cacheExp > 0 && Cache() != nil {
				Cache().Set(ctx, linkCacheKey(link), ret, cacheExp)
			}
			return ret, nil
		}
		matched := string(matches[0][1])
		ret := html.UnescapeString(matched)
		if cacheExp > 0 && Cache() != nil {
			Cache().Set(ctx, linkCacheKey(link), ret, cacheExp)
		}
		return ret, nil
	} else if strings.Contains(parsedLink.Path, "pddpage") && parsedLink.Query().Has("goodsId") {
		return goutil.StringsJoin("https://mobile.yangkeduo.com/goods.html?goods_id=", parsedLink.Query().Get("goodsId")), nil
	} else {
		resp, err := http.DefaultClient.Get(link)
		if err != nil {
			return "", errors.Join(errors.New("下载链接内容失败"), err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Join(errors.New("读取链接内容失败"), err)
		}
		regs := []string{`(https\://mobile\.yangkeduo\.com/goods\.html\?goods_id\=\d+)`}
		for _, reg := range regs {
			re := regexp.MustCompile(reg)
			matches := re.FindAll(body, 1)
			if len(matches) == 0 {
				continue
			}
			return string(matches[0]), nil
		}
	}
	return "", errors.Join(errors.New("无法识别落地页链接"), nil)
}

func ExtractPid(ctx context.Context, link string) (string, *url.URL, string, error) {
	var err error
	if link, err = ExtractLink(ctx, link, 0); err != nil {
		return "", nil, "", err
	}
	parsedLink, err := url.ParseRequestURI(link)
	if err != nil {
		return "", nil, "", errors.Join(errors.New("解析链接失败"), err)
	}
	if parsedLink.Host != "s.click.taobao.com" {
		page := parsedLink.Query().Get("page")
		if page == "" {
			return "", nil, "", errors.Join(errors.New("非淘宝客链接"), nil)
		}
		parsedLink, err = url.ParseRequestURI(page)
		if err != nil {
			return "", nil, "", errors.Join(errors.New("解析链接失败"), err)
		}
		if parsedLink.Host != "s.click.taobao.com" {
			return "", nil, "", errors.Join(errors.New("非淘宝客链接"), nil)
		}
	}
	tbkLink := parsedLink.String()
	oriLink, err := getTbkOriLink(tbkLink)
	if err != nil {
		return tbkLink, nil, "", err
	}
	aliTrack := oriLink.Query().Get("ali_trackid")
	parts := strings.Split(aliTrack, ":")
	if len(parts) != 3 {
		return tbkLink, oriLink, "", errors.New("invalid ali_trackid")
	}
	return tbkLink, oriLink, parts[1], err
}

func GetDeeplinkSku(ctx context.Context, link string) (uint64, Platform, error) {
	link = html.UnescapeString(link)
	parsedURL, err := url.ParseRequestURI(link)
	if err != nil {
		return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("解析链接失败"), err)
	}
	scheme := strings.ToLower(parsedURL.Scheme)
	switch scheme {
	case "tbopen", "taobao":
		if h5Page, err := url.ParseRequestURI(parsedURL.Query().Get("h5Url")); err != nil {
			return 0, UNKNOWN_PLATFORM, err
		} else if itemID, _ := GetTaobaoItemIDFromLink(h5Page); itemID > 0 {
			return itemID, TAOBAO, nil
		}
	case "openjd", "openapp.jdmobile":
		params := parsedURL.Query().Get("params")
		if params == "" {
			return 0, UNKNOWN_PLATFORM, errors.New("京东直达链接参数错误")
		}
		var decParam struct {
			Url string `json:"url,omitempty"`
		}
		if err := json.Unmarshal([]byte(params), &decParam); err != nil || decParam.Url == "" {
			return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("京东直达链接参数错误"), err)
		}
		if parsedPage, err := url.ParseRequestURI(decParam.Url); err != nil {
			return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("解析京东链接失败"), err)
		} else if itemID := GetJDItemIDFromLink(parsedPage); itemID > 0 {
			return itemID, JD, nil
		}
	case "pddopen", "pinduoduo":
		if h5Page, err := url.ParseRequestURI(parsedURL.Query().Get("h5Url")); err != nil {
			return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("解析拼多多链接失败"), err)
		} else if itemID, _ := strconv.ParseUint(h5Page.Query().Get("goods_id"), 10, 64); itemID > 0 {
			return itemID, PDD, nil
		}
	case "imeituan":
		query := parsedURL.Query()
		subLink := query.Get("targetPath")
		if subLink == "" {
			subLink = query.Get("url")
		}
		if subLink != "" {
			if h5Page, err := url.ParseRequestURI(subLink); err != nil {
				return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("解析美团链接失败"), err)
			} else if itemID, _ := strconv.ParseUint(h5Page.Query().Get("sku_id"), 10, 64); itemID > 0 {
				return itemID, MEITUAN, nil
			}
		}
	default:
		return 0, UNKNOWN_PLATFORM, errors.New("未知平台链接")
	}
	return 0, UNKNOWN_PLATFORM, errors.New("无法获取商品ID")
}

func GetLinkSku(ctx context.Context, link string) (uint64, Platform, error) {
	link = html.UnescapeString(link)
	link, err := ExtractLink(ctx, link, 0)
	if err != nil {
		return 0, UNKNOWN_PLATFORM, err
	}
	parsedUrl, err := url.ParseRequestURI(link)
	if err != nil {
		return 0, UNKNOWN_PLATFORM, errors.Join(errors.New("解析链接失败"), err)
	}
	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return GetDeeplinkSku(ctx, link)
	}
	query := parsedUrl.Query()
	if parsedUrl.Host == "xhsh.xibao100.com" {
		if query.Get("page") != "" {
			return GetLinkSku(ctx, query.Get("page"))
		}
		if strings.HasPrefix(parsedUrl.Path, "/i/") {
			linkPath := strings.Trim(parsedUrl.Path, "/")
			parts := strings.Split(linkPath, "/")
			if itemId, _ := strconv.ParseUint(parts[3], 10, 64); itemId > 0 {
				return itemId, UNKNOWN_PLATFORM, nil
			} else if parts := strings.Split(linkPath, "-"); len(parts) == 0 {
				if arr := goutil.DecodeUint64s(parts[1]); len(arr) > 1 && arr[1] > 0 {
					return arr[1], UNKNOWN_PLATFORM, nil
				}
			}
		}
	} else if parsedUrl.Host == "wxmall.xibao100.com" {
		if itemId, _ := strconv.ParseUint(query.Get("id"), 10, 64); itemId > 0 {
			return itemId, WECHAT, nil
		}
		if itemId, _ := strconv.ParseUint(query.Get("sku_id"), 10, 64); itemId > 0 {
			return itemId, WECHAT, nil
		}
	} else if strings.HasSuffix(parsedUrl.Host, ".taobao.com") || strings.HasSuffix(parsedUrl.Host, ".tmall.com") || strings.HasSuffix(parsedUrl.Host, ".tb.cn") || strings.HasSuffix(parsedUrl.Host, ".duanqu.com") {
		if itemId, _ := GetTaobaoItemIDFromLink(parsedUrl); itemId > 0 {
			return itemId, TAOBAO, nil
		}
	} else if strings.HasSuffix(parsedUrl.Host, ".jd.com") || strings.HasSuffix(parsedUrl.Host, ".jd.hk") {
		if itemId := GetJDItemIDFromLink(parsedUrl); itemId > 0 {
			return itemId, JD, nil
		}
	} else if strings.HasSuffix(parsedUrl.Host, ".pinduoduo.com") || strings.HasSuffix(parsedUrl.Host, ".yangkeduo.com") {
		if itemId, _ := strconv.ParseUint(parsedUrl.Query().Get("goods_id"), 10, 64); itemId > 0 {
			return itemId, PDD, nil
		}
	} else if strings.HasSuffix(parsedUrl.Host, ".meituan.com") {
		if itemId, _ := GetMeituanItemIDFromLink(parsedUrl); itemId > 0 {
			return itemId, MEITUAN, nil
		}
	}
	return 0, UNKNOWN_PLATFORM, errors.New(goutil.StringsJoin("无法获取商品ID, 链接:", link))
}

func GetJDItemIDFromLink(parsedUrl *url.URL) uint64 {
	switch parsedUrl.Host {
	case "jkgj-isv.isvjcloud.com":
		if parsedUrl.Query().Get("url") != "" {
			if parsedLink, err := url.ParseRequestURI(parsedUrl.Query().Get("url")); err == nil {
				return GetJDItemIDFromLink(parsedLink)
			} else {
				return 0
			}
		} else if parsedUrl.Path == "/ad/user/activity" {
			itemId, _ := strconv.ParseUint(parsedUrl.Query().Get("item_id"), 10, 64)
			return itemId
		}
	case "platform.m.jd.com":
		if parsedLink, err := url.ParseRequestURI(parsedUrl.Query().Get("spreadUrl")); err == nil {
			return GetJDItemIDFromLink(parsedLink)
		}
	case "pro.m.jd.com":
		resp, err := http.DefaultClient.Get(parsedUrl.String())
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}
		regs := []string{`//item\.m\.jd\.com/ware/view\.action\?wareId\=(\d+)`, `//item\.m\.jd\.com/product/(\d+)\.html`, `//item\.jd\.com/(\d+)\.html`, `//item\.yiyaojd\.com/(\d+)\.html`, `"(https://u\.jd\.com/\w+)"`}
		for idx, reg := range regs {
			re := regexp.MustCompile(reg)
			matches := re.FindAllSubmatch(body, 1)
			if len(matches) == 0 || len(matches[0]) != 2 {
				continue
			}
			if idx == 3 {
				if parsedLink, err := url.ParseRequestURI(string(matches[0][1])); err == nil {
					return GetJDItemIDFromLink(parsedLink)
				}
			} else if itemId, _ := strconv.ParseUint(string(matches[0][1]), 10, 64); itemId > 0 {
				return itemId
			}
		}
		return 0
	case "u.jd.com", "union-click.jd.com":
		if parsedUrl.Path == "/jdc" {
			parsedUrl.Path = "/jda"
		}
		if itemId, _ := strconv.ParseUint(parsedUrl.Query().Get("wareId"), 10, 64); itemId > 0 {
			return itemId
		}
		httpReq, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
		if err != nil {
			return 0
		}
		httpReq.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
		httpResp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return 0
		}
		defer httpResp.Body.Close()
		query := httpResp.Request.URL.Query()
		if httpResp.Request.URL.Host == "trade.m.jd.com" && query.Get("referer") != "" {
			if parsedUrl, err := url.ParseRequestURI(query.Get("referer")); err == nil {
				return GetJDItemIDFromLink(parsedUrl)
			}
		} else if httpResp.Request.URL.Host == "item.m.jd.com" {
			return GetJDItemIDFromLink(httpResp.Request.URL)
		} else if httpResp.Request.URL.Host == "pro.m.jd.com" {
			body, _ := io.ReadAll(httpResp.Body)
			regs := []string{`//item\.m\.jd\.com/ware/view\.action\?wareId\=(\d+)`, `//item\.m\.jd\.com/product/(\d+)\.html`, `//item\.jd\.com/(\d+)\.html`, `//item\.yiyaojd.com/(\d+).html`}
			for _, reg := range regs {
				re := regexp.MustCompile(reg)
				matches := re.FindAllSubmatch(body, 1)
				if len(matches) == 0 || len(matches[0]) != 2 {
					continue
				}
				if itemId, _ := strconv.ParseUint(string(matches[0][1]), 10, 64); itemId > 0 {
					return itemId
				}
			}
		} else if redt := query.Get("returnurl"); redt != "" {
			if parsedUrl, err := url.ParseRequestURI(redt); err == nil {
				return GetJDItemIDFromLink(parsedUrl)
			}
		} else {
			body, _ := io.ReadAll(httpResp.Body)
			regs := []string{`'(?U)(https://u\.jd\.com/jda\?.+)'`, `(?U)(//item\.jd\.com/\d+\.html)`, `(?U)(//item\.m\.jd\.com/ware/view\.action\?wareId\=\d+)`, `(?U)(//item\.m\.jd\.com/product/\d+\.html)`}
			for _, reg := range regs {
				re := regexp.MustCompile(reg)
				match := re.FindAllSubmatch(body, 1)
				if len(match) > 0 && len(match[0]) > 1 {
					if parsedUrl, err := url.ParseRequestURI(string(match[0][1])); err == nil {
						return GetJDItemIDFromLink(parsedUrl)
					}
				}
			}
		}
	default:
		if !strings.HasSuffix(parsedUrl.Host, ".jd.com") && !strings.HasSuffix(parsedUrl.Host, ".jd.hk") && !strings.HasSuffix(parsedUrl.Host, ".yiyaojd.com") {
			return 0
		}
		re := regexp.MustCompile(`(\d+)\.html`)
		if match := re.FindAllStringSubmatch(parsedUrl.Path, 1); len(match) == 1 && len(match[0]) == 2 {
			itemId, _ := strconv.ParseUint(match[0][1], 10, 64)
			return itemId
		} else if itemId, _ := strconv.ParseUint(parsedUrl.Query().Get("wareId"), 10, 64); itemId > 0 {
			return itemId
		} else if redt := parsedUrl.Query().Get("to"); redt != "" {
			if parsedRedt, err := url.ParseRequestURI(redt); err == nil {
				return GetJDItemIDFromLink(parsedRedt)
			}
		}
	}
	return 0
}

func GetTaobaoItemIDFromLink(parsedUrl *url.URL) (uint64, error) {
	switch parsedUrl.Host {
	case "login.taobao.com":
		if redt := parsedUrl.Query().Get("redirectURL"); redt != "" {
			if parsedRedt, err := url.ParseRequestURI(redt); err != nil {
				return 0, err
			} else {
				return GetTaobaoItemIDFromLink(parsedRedt)
			}
		}
	case "login.1688.com":
		if redt := parsedUrl.Query().Get("target"); redt != "" {
			if parsedRedt, err := url.ParseRequestURI(redt); err != nil {
				return 0, err
			} else {
				return GetTaobaoItemIDFromLink(parsedRedt)
			}
		}
	case "s.click.taobao.com", "uland.taobao.com":
		for i := 0; i < 2; i++ {
			var query url.Values
			if i == 0 {
				query = parsedUrl.Query()
			} else {
				tbkLink, err := getTbkOriLink(parsedUrl.String())
				if err != nil {
					return 0, err
				}
				query = tbkLink.Query()
			}
			if itemId, _ := strconv.ParseUint(query.Get("itemId"), 10, 64); itemId > 0 {
				return itemId, nil
			} else if itemId, _ := strconv.ParseUint(query.Get("item_id"), 10, 64); itemId > 0 {
				return itemId, nil
			} else if itemId, _ := strconv.ParseUint(query.Get("id"), 10, 64); itemId > 0 {
				return itemId, nil
			}
		}
	case "mo.m.tmall.com", "mo.m.taobao.com":
		for i := 0; i < 2; i++ {
			var query url.Values
			if i == 0 {
				query = parsedUrl.Query()
			} else {
				tbkLink, err := getTbkOriLink(parsedUrl.String())
				if err != nil {
					fmt.Println(err)
					return 0, err
				}
				query = tbkLink.Query()
			}
			if itemId, _ := strconv.ParseUint(query.Get("itemId"), 10, 64); itemId > 0 {
				return itemId, nil
			} else if itemId, _ := strconv.ParseUint(query.Get("item_id"), 10, 64); itemId > 0 {
				return itemId, nil
			} else if itemId, _ := strconv.ParseUint(query.Get("id"), 10, 64); itemId > 0 {
				return itemId, nil
			}
		}
	case "m.duanqu.com":
		if parsedUrl.Query().Get("query") != "" {
			if query, err := url.ParseQuery(parsedUrl.Query().Get("query")); err != nil {
				return 0, err
			} else if query.Get("promo_id") != "1" && query.Get("ews_act_type") != "" {
				if iid, err := strconv.ParseUint(query.Get("goodsId"), 10, 64); err == nil && iid > 1 {
					return iid, nil
				}
			} else if query.Get("redt") != "" {
				if iid, err := strconv.ParseUint(query.Get("goodsId"), 10, 64); err == nil && iid > 1 {
					return iid, nil
				}
			}
		}
	case "gateway.alihealth.taobao.com":
		httpReq, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
		if err != nil {
			return 0, err
		}
		httpReq.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
		httpResp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return 0, err
		}
		defer httpResp.Body.Close()
		if httpResp.Request.URL.Host == "detail.m.tmall.com" {
			return GetTaobaoItemIDFromLink(httpResp.Request.URL)
		}
	default:
		if !strings.HasSuffix(parsedUrl.Host, ".taobao.com") && !strings.HasSuffix(parsedUrl.Host, ".tmall.com") && !strings.HasSuffix(parsedUrl.Host, ".tb.cn") && !strings.HasSuffix(parsedUrl.Host, ".tmall.hk") {
			return 0, errors.New("非淘宝链接")
		}
		query := parsedUrl.Query()
		if itemId, _ := strconv.ParseUint(query.Get("id"), 10, 64); itemId > 0 {
			return itemId, nil
		}
	}
	return 0, errors.New("not found")
}

func GetMeituanItemIDFromLink(parsedURL *url.URL) (uint64, error) {
	query := parsedURL.Query()
	if str := query.Get("page_sku_id"); str != "" {
		if itemID, _ := strconv.ParseUint(str, 10, 64); itemID > 0 {
			return itemID, nil
		}
	}
	if str := query.Get("sku_id"); str != "" {
		if itemID, _ := strconv.ParseUint(str, 10, 64); itemID > 0 {
			return itemID, nil
		}
	}
	if str := query.Get("deepLinkUrl"); str != "" {
		if itemID, platform, _ := GetDeeplinkSku(context.Background(), str); itemID > 0 && platform == MEITUAN {
			return itemID, nil
		}
	}
	return 0, errors.New("not found")
}

func getTbkOriLink(link string) (*url.URL, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.Join(errors.New("初始化cookiejar失败"), err)
	}
	httpClient := &http.Client{
		Jar: jar,
	}
	unionUrl, itemID, err := getTbkRedirectLink(httpClient, link)
	if err != nil {
		return nil, err
	}
	if itemID > 0 {
		return unionUrl, nil
	}
	redirectURL := unionUrl.String()
	httpReq, err := http.NewRequest(http.MethodGet, redirectURL, nil)
	if err != nil {
		return nil, errors.Join(errors.New("初始化http request失败"), err)
	}
	httpReq.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	httpReq.Header.Add("Referer", redirectURL)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Join(errors.New("下载链接内容失败"), err)
	}
	defer httpResp.Body.Close()
	ret := httpResp.Request.URL
	query := ret.Query()
	if itemId, _ := strconv.ParseUint(query.Get("itemId"), 10, 64); itemId > 0 {
		return ret, nil
	} else if itemId, _ := strconv.ParseUint(query.Get("item_id"), 10, 64); itemId > 0 {
		return ret, nil
	} else if itemId, _ := strconv.ParseUint(query.Get("id"), 10, 64); itemId > 0 {
		return ret, nil
	}
	doc, err := goquery.NewDocumentFromReader(httpResp.Body)
	if err != nil {
		return nil, errors.Join(errors.New("解析html失败"), err)
	}
	var found bool
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if found {
			return
		}
		if link, exists := s.Attr("href"); !exists {
			return
		} else if parsedURL, err := url.ParseRequestURI(link); err == nil {
			query := parsedURL.Query()
			if query.Has("itemId") || query.Has("item_id") {
				ret = parsedURL
				found = true
			}
		}
	})
	if !found {
		doc.Find("div").Each(func(i int, s *goquery.Selection) {
			if found {
				return
			}
			if itemIDAttr, exists := s.Attr("item_id"); !exists {
				return
			} else if itemID, err := strconv.ParseUint(itemIDAttr, 10, 64); err == nil && itemID > 0 {
				ret, _ = url.ParseRequestURI(goutil.StringsJoin("https://h5.m.taobao.com/awp/core/detail.htm?id=", itemIDAttr))
				found = true
			}
		})
	}
	return ret, nil
}

func getTbkRedirectLink(clt *http.Client, link string) (*url.URL, uint64, error) {
	httpReq, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, 0, errors.Join(errors.New("初始化http request失败"), err)
	}
	httpReq.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")
	httpResp, err := clt.Do(httpReq)
	if err != nil {
		return nil, 0, errors.Join(errors.New("下载链接内容失败"), err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, 0, errors.Join(errors.New("读取链接内容失败"), err)
	}
	re := regexp.MustCompile(`var real_jump_address = '(.+)'`)
	match := re.FindAllSubmatch(body, 1)
	if len(match) != 1 || len(match[0]) != 2 {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return nil, 0, errors.Join(errors.New("解析链接内容失败"), err)
		}
		var (
			ret    *url.URL
			itemID uint64
		)
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if itemID > 0 {
				return
			}
			if link, exists := s.Attr("href"); !exists {
				return
			} else if parsedURL, err := url.ParseRequestURI(link); err == nil {
				query := parsedURL.Query()
				if query.Has("itemId") {
					itemID, _ = strconv.ParseUint(query.Get("itemId"), 10, 64)
				} else if query.Has("item_id") {
					itemID, _ = strconv.ParseUint(query.Get("item_id"), 10, 64)
				}
				if itemID > 0 {
					ret = parsedURL
				}
			}
		})
		if itemID == 0 {
			doc.Find("div").Each(func(i int, s *goquery.Selection) {
				if itemID > 0 {
					return
				}
				if itemIDAttr, exists := s.Attr("item_id"); !exists {
					return
				} else if itemID, err = strconv.ParseUint(itemIDAttr, 10, 64); err == nil && itemID > 0 {
					ret, _ = url.ParseRequestURI(goutil.StringsJoin("https://h5.m.taobao.com/awp/core/detail.htm?id=", itemIDAttr))
				}
			})
		}
		if itemID > 0 {
			return ret, itemID, nil
		}
		return nil, 0, errors.New(goutil.StringsJoin("无法获取淘宝客跳转链接:", link))
	}
	ret, err := url.ParseRequestURI(html.UnescapeString(string(match[0][1])))
	if err != nil {
		return nil, 0, errors.New("解析链接失败")
	}
	return ret, 0, nil
}
