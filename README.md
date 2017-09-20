# gowiki

* TODO
  * main.css 編集 (body, h1 ...)
    * viewtree: 表示を見やすく + 前回の開いている画面を保存
    * diff 表示画面色付け & diff 表示を見やすくする
    * delete: 確認画面表示
  * Docker, nginx init
  * render 関数整理

* Option
  * 目次表示
    * 項目ごとに分割
    * 項目ごとに編集できるようにする
      * https://github.com/russross/blackfriday
      * http://step-learn.com/markdown/20141017-idx.html
      * http://qiita.com/Qiita/items/c686397e4a0f4f11683d
    * サイドバー作成
      * https://getbootstrap.com/docs/4.0/examples/
  * 検索 (header: search)
  * リアルタイムプレビュー 
  * /repo /repo/... のルーティングを統一, subdirectory 設定
  * git の初期設定


* Memo
  * content が "" の show 画面表示の際に redirect をする > log: multiple response.WriteHeader calls
  * test.go 作成
    * golint .
