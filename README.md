# online-chat-messenger
クライアント・サーバー方式のリアルタイムチャットアプリケーションです。

## 特徴
- ### TCP
  チャットルームの作成および参加リクエストにはTCP接続を利用しています。

- ### UDP
  チャットルーム参加後のメッセージはUDPを利用して送信されます。

- ### カスタムプロトコル
  TCPとUDPの両方で、独自のアプリケーション層プロトコルを設計して利用しています。

## 機能
ユーザーはCLIツールから以下の機能を利用できます。
- チャットルームの作成
- チャットルームへの参加
- チャット

## こだわった点
カスタムプロトコルにstateの項目を用意しました。
stateには、
- サーバーがリクエストを正しく処理したのか(success)
- サーバー側の原因で処理できなかったのか(fail)
- リクエストが不正なもので処理できなかったのか(invalid)
といった内容を含めることで、
CLI側でサーバーからの応答に応じたエラーハンドリングを実装しました。

例) チャットルームへの参加リクエストに含まれたパスワードが不正だったとき

  -> サーバーからstate = invalid なレスポンスを送る
  
  -> ユーザーにリクエストが拒絶されたことを通知して終了
