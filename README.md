# gowiki

* TODO
  * git commit (header: edit)
  * config file 作成
  * test.go 作成
    * golint .
  * baseurl の自動設定

* Add
  * sidebar: 折りたたみ
  * main.css 編集 (body, h1 ...)
  * 目次表示
    * 項目ごとに分割
    * 項目ごとに編集できるようにする
      * https://github.com/russross/blackfriday
      * http://step-learn.com/markdown/20141017-idx.html
      * http://qiita.com/Qiita/items/c686397e4a0f4f11683d
    * サイドバー作成
      * https://getbootstrap.com/docs/4.0/examples/

* Option
  * 差分: git diff (header: diff)
  * 検索 (header: search)
  * 画像アップロード (footer: upload)
  * リアルタイムプレビュー 
  * /repo を / に移行


* Memo
  * content が "" の show 画面表示の際に redirect をする > log: multiple response.WriteHeader calls
