import ballerina/grpc;
import ballerina/protobuf;
import ballerina/protobuf.types.'any;

public const string TYPES_V1_DESC = "0A0E74797065735F76312E70726F746F1204637275641A19676F6F676C652F70726F746F6275662F616E792E70726F746F22320A044B696E6412140A056D616A6F7218012001280952056D616A6F7212140A056D696E6F7218022001280952056D696E6F7222740A0E54696D65426173656456616C7565121C0A09737461727454696D651801200128095209737461727454696D6512180A07656E6454696D651802200128095207656E6454696D65122A0A0576616C756518032001280B32142E676F6F676C652E70726F746F6275662E416E79520576616C75652294010A0C52656C6174696F6E7368697012280A0F72656C61746564456E746974794964180120012809520F72656C61746564456E746974794964121C0A09737461727454696D651802200128095209737461727454696D6512180A07656E6454696D651803200128095207656E6454696D65120E0A0269641804200128095202696412120A046E616D6518052001280952046E616D6522DB040A06456E74697479120E0A02696418012001280952026964121E0A046B696E6418022001280B320A2E637275642E4B696E6452046B696E6412180A0763726561746564180320012809520763726561746564121E0A0A7465726D696E61746564180420012809520A7465726D696E6174656412280A046E616D6518052001280B32142E637275642E54696D65426173656456616C756552046E616D6512360A086D6574616461746118062003280B321A2E637275642E456E746974792E4D65746164617461456E74727952086D65746164617461123C0A0A6174747269627574657318072003280B321C2E637275642E456E746974792E41747472696275746573456E747279520A6174747269627574657312450A0D72656C6174696F6E736869707318082003280B321F2E637275642E456E746974792E52656C6174696F6E7368697073456E747279520D72656C6174696F6E73686970731A510A0D4D65746164617461456E74727912100A036B657918012001280952036B6579122A0A0576616C756518022001280B32142E676F6F676C652E70726F746F6275662E416E79520576616C75653A0238011A570A0F41747472696275746573456E74727912100A036B657918012001280952036B6579122E0A0576616C756518022001280B32182E637275642E54696D65426173656456616C75654C697374520576616C75653A0238011A540A1252656C6174696F6E7368697073456E74727912100A036B657918012001280952036B657912280A0576616C756518022001280B32122E637275642E52656C6174696F6E73686970520576616C75653A02380122420A1254696D65426173656456616C75654C697374122C0A0676616C75657318012003280B32142E637275642E54696D65426173656456616C7565520676616C75657322610A1152656164456E7469747952657175657374120E0A0269641801200128095202696412240A06656E7469747918022001280B320C2E637275642E456E746974795206656E7469747912160A066F757470757418032003280952066F7574707574221A0A08456E746974794964120E0A02696418012001280952026964224B0A13557064617465456E7469747952657175657374120E0A0269641801200128095202696412240A06656E7469747918022001280B320C2E637275642E456E746974795206656E7469747922070A05456D70747922360A0A456E746974794C69737412280A08656E74697469657318012003280B320C2E637275642E456E746974795208656E746974696573328F020A0B4372756453657276696365122A0A0C437265617465456E74697479120C2E637275642E456E746974791A0C2E637275642E456E7469747912330A0A52656164456E7469747912172E637275642E52656164456E74697479526571756573741A0C2E637275642E456E7469747912390A0C52656164456E74697469657312172E637275642E52656164456E74697479526571756573741A102E637275642E456E746974794C69737412370A0C557064617465456E7469747912192E637275642E557064617465456E74697479526571756573741A0C2E637275642E456E74697479122B0A0C44656C657465456E74697479120E2E637275642E456E7469747949641A0B2E637275642E456D707479421C5A1A6C6B2F64617461666F756E646174696F6E2F637275642D617069620670726F746F33";

public isolated client class CrudServiceClient {
    *grpc:AbstractClientEndpoint;

    private final grpc:Client grpcClient;

    public isolated function init(string url, *grpc:ClientConfiguration config) returns grpc:Error? {
        self.grpcClient = check new (url, config);
        check self.grpcClient.initStub(self, TYPES_V1_DESC);
    }

    isolated remote function CreateEntity(Entity|ContextEntity req) returns Entity|grpc:Error {
        map<string|string[]> headers = {};
        Entity message;
        if req is ContextEntity {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/CreateEntity", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <Entity>result;
    }

    isolated remote function CreateEntityContext(Entity|ContextEntity req) returns ContextEntity|grpc:Error {
        map<string|string[]> headers = {};
        Entity message;
        if req is ContextEntity {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/CreateEntity", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <Entity>result, headers: respHeaders};
    }

    isolated remote function ReadEntity(ReadEntityRequest|ContextReadEntityRequest req) returns Entity|grpc:Error {
        map<string|string[]> headers = {};
        ReadEntityRequest message;
        if req is ContextReadEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/ReadEntity", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <Entity>result;
    }

    isolated remote function ReadEntityContext(ReadEntityRequest|ContextReadEntityRequest req) returns ContextEntity|grpc:Error {
        map<string|string[]> headers = {};
        ReadEntityRequest message;
        if req is ContextReadEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/ReadEntity", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <Entity>result, headers: respHeaders};
    }

    isolated remote function ReadEntities(ReadEntityRequest|ContextReadEntityRequest req) returns EntityList|grpc:Error {
        map<string|string[]> headers = {};
        ReadEntityRequest message;
        if req is ContextReadEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/ReadEntities", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <EntityList>result;
    }

    isolated remote function ReadEntitiesContext(ReadEntityRequest|ContextReadEntityRequest req) returns ContextEntityList|grpc:Error {
        map<string|string[]> headers = {};
        ReadEntityRequest message;
        if req is ContextReadEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/ReadEntities", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <EntityList>result, headers: respHeaders};
    }

    isolated remote function UpdateEntity(UpdateEntityRequest|ContextUpdateEntityRequest req) returns Entity|grpc:Error {
        map<string|string[]> headers = {};
        UpdateEntityRequest message;
        if req is ContextUpdateEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/UpdateEntity", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <Entity>result;
    }

    isolated remote function UpdateEntityContext(UpdateEntityRequest|ContextUpdateEntityRequest req) returns ContextEntity|grpc:Error {
        map<string|string[]> headers = {};
        UpdateEntityRequest message;
        if req is ContextUpdateEntityRequest {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/UpdateEntity", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <Entity>result, headers: respHeaders};
    }

    isolated remote function DeleteEntity(EntityId|ContextEntityId req) returns Empty|grpc:Error {
        map<string|string[]> headers = {};
        EntityId message;
        if req is ContextEntityId {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/DeleteEntity", message, headers);
        [anydata, map<string|string[]>] [result, _] = payload;
        return <Empty>result;
    }

    isolated remote function DeleteEntityContext(EntityId|ContextEntityId req) returns ContextEmpty|grpc:Error {
        map<string|string[]> headers = {};
        EntityId message;
        if req is ContextEntityId {
            message = req.content;
            headers = req.headers;
        } else {
            message = req;
        }
        var payload = check self.grpcClient->executeSimpleRPC("crud.CrudService/DeleteEntity", message, headers);
        [anydata, map<string|string[]>] [result, respHeaders] = payload;
        return {content: <Empty>result, headers: respHeaders};
    }
}

public type ContextEntityId record {|
    EntityId content;
    map<string|string[]> headers;
|};

public type ContextEntity record {|
    Entity content;
    map<string|string[]> headers;
|};

public type ContextEmpty record {|
    Empty content;
    map<string|string[]> headers;
|};

public type ContextUpdateEntityRequest record {|
    UpdateEntityRequest content;
    map<string|string[]> headers;
|};

public type ContextEntityList record {|
    EntityList content;
    map<string|string[]> headers;
|};

public type ContextReadEntityRequest record {|
    ReadEntityRequest content;
    map<string|string[]> headers;
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type EntityId record {|
    string id = "";
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type Entity record {|
    string id = "";
    Kind kind = {};
    string created = "";
    string terminated = "";
    TimeBasedValue name = {};
    record {|string key; 'any:Any value;|}[] metadata = [];
    record {|string key; TimeBasedValueList value;|}[] attributes = [];
    record {|string key; Relationship value;|}[] relationships = [];
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type TimeBasedValue record {|
    string startTime = "";
    string endTime = "";
    'any:Any value = {};
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type Empty record {|
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type UpdateEntityRequest record {|
    string id = "";
    Entity entity = {};
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type TimeBasedValueList record {|
    TimeBasedValue[] values = [];
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type Kind record {|
    string major = "";
    string minor = "";
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type Relationship record {|
    string relatedEntityId = "";
    string startTime = "";
    string endTime = "";
    string id = "";
    string name = "";
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type EntityList record {|
    Entity[] entities = [];
|};

@protobuf:Descriptor {value: TYPES_V1_DESC}
public type ReadEntityRequest record {|
    string id = "";
    Entity entity = {};
    string[] output = [];
|};
