# gowiki

* TODO
  * rm dir, file
    * action=D: ファイル、ディレクトリ一覧出してチェックボックスで削除
    * 確認画面
  * 画像アップロード (footer: upload)

* Add
  * sidebar: 折りたたみ
  * main.css 編集 (body, h1 ...)
  * Docker 化

* Option
  * 目次表示
    * 項目ごとに分割
    * 項目ごとに編集できるようにする
      * https://github.com/russross/blackfriday
      * http://step-learn.com/markdown/20141017-idx.html
      * http://qiita.com/Qiita/items/c686397e4a0f4f11683d
    * サイドバー作成
      * https://getbootstrap.com/docs/4.0/examples/
  * diff 表示画面色付け & diff 表示を見やすくする
  * 検索 (header: search)
  * リアルタイムプレビュー 
  * /repo /repo/... のルーティングを統一


* Memo
  * content が "" の show 画面表示の際に redirect をする > log: multiple response.WriteHeader calls
  * test.go 作成
    * golint .
