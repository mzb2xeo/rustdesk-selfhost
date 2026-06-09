use hbb_common::{
    protobuf::Message,
    rendezvous_proto::{
        register_pk_response, rendezvous_message, ConnType, PunchHoleRequest, RegisterPk,
        RegisterPkResponse, RendezvousMessage,
    },
};

#[test]
fn client_1_4_7_registration_fields_round_trip() {
    let mut message = RendezvousMessage::new();
    message.set_register_pk(RegisterPk {
        id: "123456789".to_owned(),
        uuid: vec![1, 2, 3].into(),
        pk: vec![4, 5, 6].into(),
        no_register_device: true,
        ..Default::default()
    });

    let decoded = RendezvousMessage::parse_from_bytes(&message.write_to_bytes().unwrap()).unwrap();
    let Some(rendezvous_message::Union::RegisterPk(register_pk)) = decoded.union else {
        panic!("expected register_pk");
    };
    assert!(register_pk.no_register_device);

    let response = RegisterPkResponse {
        result: register_pk_response::Result::NOT_DEPLOYED.into(),
        ..Default::default()
    };
    assert_eq!(
        response.result.enum_value().unwrap(),
        register_pk_response::Result::NOT_DEPLOYED
    );
}

#[test]
fn client_1_4_7_punch_fields_round_trip() {
    let mut message = RendezvousMessage::new();
    message.set_punch_hole_request(PunchHoleRequest {
        id: "987654321".to_owned(),
        conn_type: ConnType::TERMINAL.into(),
        udp_port: 32116,
        force_relay: true,
        upnp_port: 32117,
        socket_addr_v6: vec![7, 8, 9].into(),
        ..Default::default()
    });

    let decoded = RendezvousMessage::parse_from_bytes(&message.write_to_bytes().unwrap()).unwrap();
    let Some(rendezvous_message::Union::PunchHoleRequest(request)) = decoded.union else {
        panic!("expected punch_hole_request");
    };
    assert_eq!(request.conn_type.enum_value().unwrap(), ConnType::TERMINAL);
    assert_eq!(request.udp_port, 32116);
    assert!(request.force_relay);
    assert_eq!(request.upnp_port, 32117);
    assert_eq!(request.socket_addr_v6.as_ref(), &[7, 8, 9]);
}
