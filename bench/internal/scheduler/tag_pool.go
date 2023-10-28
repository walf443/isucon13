package scheduler

var tagPool = [...]string{
	"ライブ配信",
	"ゲーム実況",
	"生放送",
	"アドバイス",
	"初心者歓迎",
	"プロゲーマー",
	"新作ゲーム",
	"レトロゲーム",
	"RPG",
	"FPS",
	"アクションゲーム",
	"対戦ゲーム",
	"マルチプレイ",
	"シングルプレイ",
	"ゲーム解説",
	"ホラーゲーム",
	"イベント生放送",
	"新情報発表",
	"Q&Aセッション",
	"チャット交流",
	"視聴者参加",
	"音楽ライブ",
	"カバーソング",
	"オリジナル楽曲",
	"アコースティック",
	"歌配信",
	"楽器演奏",
	"ギター",
	"ピアノ",
	"バンドセッション",
	"DJセット",
	"トーク配信",
	"朝活",
	"夜ふかし",
	"日常話",
	"趣味の話",
	"語学学習",
	"お料理配信",
	"手料理",
	"レシピ紹介",
	"アート配信",
	"絵描き",
	"DIY",
	"手芸",
	"アニメトーク",
	"映画レビュー",
	"読書感想",
	"ファッション",
	"メイク",
	"ビューティー",
	"健康",
	"ワークアウト",
	"ヨガ",
	"ダンス",
	"旅行記",
	"アウトドア",
	"キャンプ",
	"ペットと一緒",
	"猫",
	"犬",
	"釣り",
	"ガーデニング",
	"テクノロジー",
	"ガジェット紹介",
	"プログラミング",
	"DIY電子工作",
	"ニュース解説",
	"歴史",
	"文化",
	"社会問題",
	"心理学",
	"宇宙",
	"科学",
	"マジック",
	"コメディ",
	"スポーツ",
	"サッカー",
	"野球",
	"バスケットボール",
	"ライフハック",
	"教育",
	"子育て",
	"ビジネス",
	"起業",
	"投資",
	"仮想通貨",
	"株式投資",
	"不動産",
	"キャリア",
	"スピリチュアル",
	"占い",
	"手相",
	"オカルト",
	"UFO",
	"都市伝説",
	"コンサート",
	"ファンミーティング",
	"コラボ配信",
	"記念配信",
	"生誕祭",
	"周年記念",
	"サプライズ",
	"椅子",
}

func GetTagPoolLength() int {
	return len(tagPool)
}