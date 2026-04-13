use futures_util::stream::TryStreamExt;
use futures::executor::block_on;
use std::sync::mpsc::channel;
use std::sync::mpsc::Sender;
use std::thread;

use zbus::{Connection, MessageStream, Message};
use zbus::message::Type as MessageType;
use zvariant::{Structure, Value};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let (tx, rx) = channel::<Message>();
    thread::spawn(move || block_on(iter_dbus_signals(tx)));
    for msg in rx {
        let body = msg.body();
        let dbody: Structure = body.deserialize()?;
        let field = dbody.fields();
        if matches!(field[0], Value::Str(_)) {
            println!("{}", dbody);
        }
    }
    Ok(())
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
        let header = msg.header();
        if header.message_type() == MessageType::Signal {
            if let Some(interface) = header.interface() {
                if interface.as_str() == "org.freedesktop.DBus.Properties" {
                    if let Some(member) = header.member() {
                        if member.as_str() == "PropertiesChanged" {
                            let _ = tx.send(msg);
                        }
                    }
                }
            }
        }
    };
    Ok(())
}
