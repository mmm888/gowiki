# gowiki

### TODO

* 見た目
  * HTML: delete 確認画面表示
  * CSS: main.css 編集 (body, h1 ...)
    * bootstrap 以外も使ってみる
    * viewtree: 表示を見やすく + 前回の開いている画面を保存
    * diff 表示画面色付け & diff 表示を見やすくする
  * JavaScript: viewtree 作成

* 内部処理
  * /repo /repo/... のルーティングを統一, subdirectory 設定
  * git の初期設定
  * render 関数整理
  * test.go 作成
    * golint .
  * error 処理
    * git 関連とか
  * Dockerfile 作成
  * blackfriday v2 使う

* Option
  * 目次表示
   * 項目ごとに分割
   * サイドバー作成
     * https://getbootstrap.com/docs/4.0/examples/
  * 項目ごとに編集できるようにする
    * https://github.com/russross/blackfriday
    * http://step-learn.com/markdown/20141017-idx.html
    * http://qiita.com/Qiita/items/c686397e4a0f4f11683d
  * 検索 (header: search)
  * リアルタイムプレビュー
