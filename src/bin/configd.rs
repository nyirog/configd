use futures_util::stream::TryStreamExt;
use futures::executor::block_on;
use std::sync::mpsc::channel;
use std::sync::mpsc::Sender;
use std::thread;
use zbus::{Connection, MessageStream, Message};

fn main() {
    let (tx, rx) = channel::<Message>();
    thread::spawn(move || block_on(iter_dbus_signals(tx)));
    for val in rx {
        println!("{}", val);
    }
}

async fn iter_dbus_signals(tx: Sender<Message>) -> zbus::fdo::Result<()> {
    let connection = Connection::session().await?;

    connection
        .call_method(
            Some("org.freedesktop.DBus"),
            "/org/freedesktop/DBus",
            Some("org.freedesktop.DBus.Monitoring"),
            "BecomeMonitor",
            &(&[] as &[&str], 0u32),
        )
        .await?;

    let mut stream = MessageStream::from(connection);
    while let Some(msg) = stream.try_next().await? {
        let _ = tx.send(msg);
    };
    Ok(())
}
